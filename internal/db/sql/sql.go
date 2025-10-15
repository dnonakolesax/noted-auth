package sql

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/consts"
)

const (
	addressLoggerKey = "address"
	sqlLoggerKey     = "sql"
)

const sqlFileExtension = ".sql"

type PGXConn struct {
	pool           *pgxpool.Pool
	requestTimeout time.Duration
	logger         *slog.Logger
}

type RDBError struct {
	Type  string
	Field string
}

func (err RDBError) Error() string {
	return err.Type + " " + err.Field
}

type PGXResponse struct {
	rows pgx.Rows
}

func NewPGXConn(config configs.RDBConfig, logger *slog.Logger) (*PGXConn, error) {
	connString := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		config.Login,
		config.Password,
		config.Address,
		config.Port,
		config.DBName)

	pgxConfig, err := pgxpool.ParseConfig(connString)

	if err != nil {
		return nil, err
	}

	pgxConfig.ConnConfig.ConnectTimeout = config.ConnTimeout

	pgxConfig.MaxConns = config.MaxConns
	pgxConfig.MinConns = config.MinConns
	pgxConfig.MaxConnLifetime = config.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = config.MaxConnIdleTime
	pgxConfig.HealthCheckPeriod = config.HealthCheckPeriod

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnTimeout)
	defer cancel()

	logger.Info("Starting pgxpool", slog.String(addressLoggerKey, config.Address))
	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		logger.Error("Error while starting pgxpool", slog.String(addressLoggerKey, config.Address),
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, fmt.Errorf("CreatePool error: %w", err)
	}
	logger.Info("Started pgxpool", slog.String(addressLoggerKey, config.Address))

	poolCtx, poolCancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer poolCancel()
	logger.Info("Trying to ping pgsql", slog.String(addressLoggerKey, config.Address))
	err = pool.Ping(poolCtx)

	if err != nil {
		logger.Error("Error while pinging pgxpool", slog.String(addressLoggerKey, config.Address),
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, fmt.Errorf("ping error: %w", err)
	}
	logger.Info("Pgsql ping success", slog.String(addressLoggerKey, config.Address))

	return &PGXConn{pool: pool, requestTimeout: config.RequestTimeout, logger: logger}, nil
}

func (pc *PGXConn) Disconnect() {
	pc.logger.Info("closing connection to pgsql")
	pc.pool.Close()
}

type PGXWorker struct {
	Conn     *PGXConn
	Requests map[string]string
	Alive    *atomic.Bool
}

func LoadSQLRequests(dirPath string) (map[string]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	sqlRequests := make(map[string]string)

	for _, file := range files {
		if file.IsDir() {
			continue // Пропускаем директории
		}

		if filepath.Ext(file.Name()) != sqlFileExtension {
			continue // Пропускаем файлы без .sql расширения
		}

		filePath := filepath.Join(dirPath, file.Name())
		content, fileErr := os.ReadFile(filePath)
		if fileErr != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name(), fileErr)
		}

		// Получаем имя файла без расширения
		fileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		sqlRequests[fileName] = string(content)
	}

	return sqlRequests, nil
}

func NewPGXWorker(conn *PGXConn, alive *atomic.Bool) (*PGXWorker, error) {
	requests := make(map[string]string)
	alive.Store(true)

	return &PGXWorker{
		Conn:     conn,
		Requests: requests,
		Alive:    alive,
	}, nil
}

func (pw *PGXWorker) Exec(ctx context.Context, sql string, args ...interface{}) error {
	timeCtx, cancel := context.WithTimeout(ctx, pw.Conn.requestTimeout)
	defer cancel()

	pw.Conn.logger.DebugContext(ctx, "executing sql", slog.String(sqlLoggerKey, sql))
	_, err := pw.Conn.pool.Exec(timeCtx, sql, args...)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		pw.Alive.Store(false)
		pw.Conn.logger.ErrorContext(ctx, "failed executing sql", slog.String(sqlLoggerKey, sql),
			slog.String(consts.ErrorLoggerKey, err.Error()))
		rdbErr := new(RDBError)
		rdbErr.Type = pgErr.Code
		rdbErr.Field = pgErr.ColumnName
		return rdbErr
	}
	pw.Conn.logger.DebugContext(ctx, "done executing sql", slog.String(sqlLoggerKey, sql))

	return nil
}

func (pw *PGXWorker) Query(ctx context.Context, sql string, args ...interface{}) (*PGXResponse, error) {
	timeCtx, cancel := context.WithTimeout(ctx, pw.Conn.requestTimeout)
	defer cancel()
	pw.Conn.logger.DebugContext(ctx, "executing sql", slog.String(sqlLoggerKey, sql))
	result, err := pw.Conn.pool.Query(timeCtx, sql, args...)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		pw.Alive.Store(false)
		pw.Conn.logger.ErrorContext(ctx, "failed executing sql", slog.String(sqlLoggerKey, sql),
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return &PGXResponse{}, RDBError{Type: pgErr.Code, Field: pgErr.ColumnName}
	}
	pw.Conn.logger.DebugContext(ctx, "done executing sql", slog.String(sqlLoggerKey, sql))

	return &PGXResponse{result}, nil
}

func (pr *PGXResponse) Next() bool {
	return pr.rows.Next()
}

func (pr *PGXResponse) Scan(dest ...any) error {
	err := pr.rows.Scan(dest...)
	if err != nil {
		return fmt.Errorf("scan error: %w", err)
	}

	return nil
}

func (pr *PGXResponse) Close() error {
	pr.rows.Close()
	// Err() on the returned Rows must be checked after the Rows is closed to determine
	// if the query executed successfully as some errors can only be detected by reading the entire response.
	// e.g. A divide by zero error on the last row.
	err := pr.rows.Err()
	if err != nil {
		return err
	}
	return nil
}
