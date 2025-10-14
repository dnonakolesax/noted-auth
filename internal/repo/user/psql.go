package repo

import (
	"context"
	"errors"
	"log/slog"

	dbsql "github.com/dnonakolesax/noted-auth/internal/db/sql"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

const thisDomainName = "user"

const (
	getUserFileName = "get_user"
)

type UserRepo struct {
	worker  *dbsql.PGXWorker
	realmID string
	logger  *slog.Logger
}

func NewUserRepo(worker *dbsql.PGXWorker, realmID string, requestsPath string, logger *slog.Logger) (*UserRepo, error) {
	userRequests, err := dbsql.LoadSQLRequests(requestsPath + thisDomainName)

	if err != nil {
		logger.Error("Error loading SQL requests", slog.String("error", err.Error()))
		return nil, err
	}

	for key, value := range userRequests {
		worker.Requests[key] = value
	}

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
		ur.logger.ErrorContext(ctx, "Error executing query", slog.String("error", err.Error()))
		return model.User{}, err
	}

	if !result.Next() {
		ur.logger.WarnContext(ctx, "User not found", slog.String("id", userID))
		return model.User{}, errors.New("not found")
	}
	var user model.User
	err = result.Scan(&user.Login, &user.FirstName, &user.LastName)
	if err != nil {
		ur.logger.ErrorContext(ctx, "Error scanning row", slog.String("error", err.Error()))
		return model.User{}, err
	}

	if result.Next() {
		ur.logger.ErrorContext(ctx, "Too many rows", slog.String("id", userID))
		return model.User{}, errors.New("too many rows")
	}

	err = result.Close()
	if err != nil {
		ur.logger.ErrorContext(ctx, "Error closing result", slog.String("error", err.Error()))
		return model.User{}, err
	}
	return user, nil
}
