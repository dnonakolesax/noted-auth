package repo

import (
	"maps"
	"context"
	"errors"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	dbsql "github.com/dnonakolesax/noted-auth/internal/db/sql"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

const thisDomainName = "user"

const userIDKey = "user_id"
const userLoginKey = "user_login"

const (
	getUserFileName = "get_user"
	getUserByNameFileName = "get_user_by_name"
)

type UserRepo struct {
	worker  *dbsql.PGXWorker
	realmID string
	logger  *slog.Logger
}

func NewUserRepo(worker *dbsql.PGXWorker, realmID string, requestsPath string, logger *slog.Logger) (*UserRepo, error) {
	userRequests, err := dbsql.LoadSQLRequests(requestsPath + thisDomainName)

	if err != nil {
		logger.Error("Error loading SQL requests", slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, err
	}

	maps.Copy(worker.Requests, userRequests)

	return &UserRepo{
		worker:  worker,
		realmID: realmID,
		logger:  logger,
	}, nil
}

func (ur *UserRepo) GetUser(ctx context.Context, userID string) (model.User, error) {
	ur.logger.InfoContext(ctx, "About to execute query", slog.String("query_name", ur.worker.Requests[getUserFileName]))
	result, err := ur.worker.Query(ctx, ur.worker.Requests[getUserFileName], userID, ur.realmID)

	if err != nil {
		ur.logger.ErrorContext(ctx, "Error executing query", slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.User{}, err
	}

	if !result.Next() {
		ur.logger.WarnContext(ctx, "User not found", slog.String(userIDKey, userID))
		return model.User{}, errors.New("not found")
	}
	var user model.User
	err = result.Scan(&user.Login, &user.FirstName, &user.LastName)
	if err != nil {
		ur.logger.ErrorContext(ctx, "Error scanning row", slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.User{}, err
	}

	if result.Next() {
		ur.logger.ErrorContext(ctx, "Too many rows", slog.String(userIDKey, userID))
		return model.User{}, errors.New("too many rows")
	}

	err = result.Close()
	if err != nil {
		ur.logger.ErrorContext(ctx, "Error closing result", slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.User{}, err
	}
	return user, nil
}

func (ur *UserRepo) IDByName(ctx context.Context, login string) (model.UserID, error) {
	ur.logger.InfoContext(ctx, "About to execute query", slog.String("query_name", ur.worker.Requests[getUserByNameFileName]))
	result, err := ur.worker.Query(ctx, ur.worker.Requests[getUserByNameFileName], login, ur.realmID)

	if err != nil {
		ur.logger.ErrorContext(ctx, "Error executing query", slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.UserID{}, err
	}

	if !result.Next() {
		ur.logger.WarnContext(ctx, "User not found", slog.String(userLoginKey, login))
		return model.UserID{}, errors.New("not found")
	}
	var user model.UserID
	err = result.Scan(&user.ID)
	if err != nil {
		ur.logger.ErrorContext(ctx, "Error scanning row", slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.UserID{}, err
	}

	if result.Next() {
		ur.logger.ErrorContext(ctx, "Too many rows", slog.String(userLoginKey, login))
		return model.UserID{}, errors.New("too many rows")
	}

	err = result.Close()
	if err != nil {
		ur.logger.ErrorContext(ctx, "Error closing result", slog.String(consts.ErrorLoggerKey, err.Error()))
		return model.UserID{}, err
	}
	return user, nil
}
