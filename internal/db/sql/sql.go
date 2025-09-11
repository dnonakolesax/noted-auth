package sql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGXConn struct {
	pool           *pgxpool.Pool
	requestTimeout time.Duration
}

type RDBErr struct {
	Type  string
	Field string
}

func (err RDBErr) Error() string {
	return err.Type + " " + err.Field
}

type PGXResponse struct {
	rows pgx.Rows
}

func NewPGXConn(config configs.RDBConfig) (*PGXConn, error) {
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

	pgxConfig.MaxConns = int32(config.MaxConns)
	pgxConfig.MinConns = int32(config.MinConns)
	pgxConfig.MaxConnLifetime = config.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = config.MaxConnIdleTime
	pgxConfig.HealthCheckPeriod = config.HealthCheckPeriod

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnTimeout*2)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("CreatePool error: %v", err)
	}

	poolCtx, poolCancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer poolCancel()
	err = pool.Ping(poolCtx)

	if err != nil {
		return nil, fmt.Errorf("Ping error: %v", err)
	}

	return &PGXConn{pool: pool, requestTimeout: config.RequestTimeout}, nil
}

func (pc *PGXConn) Disconnect() {
	pc.pool.Close()
}

type PGXWorker struct {
	Conn     *PGXConn
	Requests map[string]string
}

func LoadSQLRequests(dirPath string) (map[string]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	sqlRequests := make(map[string]string)

	for _, file := range files {
		if file.IsDir() {
			continue // Пропускаем директории
		}

		if filepath.Ext(file.Name()) != ".sql" {
			continue // Пропускаем файлы без .sql расширения
		}

		filePath := filepath.Join(dirPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file.Name(), err)
		}

		// Получаем имя файла без расширения
		fileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		sqlRequests[fileName] = string(content)
	}

	return sqlRequests, nil
}

func NewPGXWorker(conn *PGXConn) (*PGXWorker, error) {
	//requests, err := LoadSQLRequests("./internal/db/sql/requests")

	requests := make(map[string]string)
	// if err != nil {
	// 	return nil, err
	// }

	return &PGXWorker{
		Conn:     conn,
		Requests: requests,
	}, nil
}

func (pw *PGXWorker) Exec(ctx context.Context, sql string, args ...interface{}) error {
	timeCtx, cancel := context.WithTimeout(ctx, pw.Conn.requestTimeout)
	defer cancel()
	_, err := pw.Conn.pool.Exec(timeCtx, sql, args...)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		rdbErr := new(RDBErr)
		rdbErr.Type = pgErr.Code
		rdbErr.Field = pgErr.ColumnName
		return rdbErr
	}

	return nil
}

func (pw *PGXWorker) Query(ctx context.Context, sql string, args ...interface{}) (*PGXResponse, error) {
	timeCtx, cancel := context.WithTimeout(ctx, pw.Conn.requestTimeout)
	defer cancel()
	result, err := pw.Conn.pool.Query(timeCtx, sql, args...)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return &PGXResponse{}, RDBErr{Type: pgErr.Code, Field: pgErr.ColumnName}
	}

	return &PGXResponse{result}, nil
}

func (pw *PGXWorker) Transaction(request []string) error {
	if len(request) > 0 {
		return fmt.Errorf("unimplemented")
	}
	return nil
}

func (pr *PGXResponse) Next() bool {
	return pr.rows.Next()
}

func (pr *PGXResponse) Scan(dest ...any) error {

	// for pr.rows.Next() {
	err := pr.rows.Scan(dest...)
	if err != nil {
		return fmt.Errorf("scan error: %v", err)
	}
	//}

	return nil
}

func (pr *PGXResponse) Close() error {
	pr.rows.Close()
	//Err() on the returned Rows must be checked after the Rows is closed to determine if the query executed successfully as some errors can only be detected by reading the entire response.
	//e.g. A divide by zero error on the last row.
	err := pr.rows.Err()
	if err != nil {
		return err
	}
	return nil
}
