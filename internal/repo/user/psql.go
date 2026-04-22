package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	dbsql "github.com/dnonakolesax/noted-auth/internal/db/sql"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

const thisDomainName = "user"

const (
	userIDKey     = "user_id"
	userLoginKey  = "user_login"
	userPrefixKey = "user_prefix"
)

const (
	getUserFileName        = "get_user"
	getUserByNameFileName  = "get_user_by_name"
	searchByPrefixFileName = "search_by_prefix"
)

type Repo struct {
	worker   dbsql.IPGXWorker
	realmID  string
	logger   *slog.Logger
	requests map[string]string
}

func NewUserRepo(worker dbsql.IPGXWorker, realmID string, requestsPath string, logger *slog.Logger) (*Repo, error) {
	userRequests, err := dbsql.LoadSQLRequests(requestsPath + thisDomainName)

	if err != nil {
		logger.Error("Error loading SQL requests", slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, err
	}

	return &Repo{
		worker:   worker,
		realmID:  realmID,
		logger:   logger,
		requests: userRequests,
	}, nil
}

func (ur *Repo) GetUser(ctx context.Context, userID string) (model.User, error) {
	ur.logger.InfoContext(ctx, "About to execute query", slog.String("query_name", ur.requests[getUserFileName]))
	result, err := ur.worker.Query(ctx, ur.requests[getUserFileName], userID, ur.realmID)

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

// SearchByPrefix returns up to `limit` users whose usernames start with the given prefix.
// The prefix is matched case-sensitively using SQL LIKE; any LIKE wildcards (% and _)
// inside the prefix are escaped so they are treated as literal characters.
func (ur *Repo) SearchByPrefix(ctx context.Context, prefix string, limit int) ([]model.UserSuggestion, error) {
	pattern := escapeLikePattern(prefix) + "%"

	ur.logger.InfoContext(ctx, "About to execute query",
		slog.String("query_name", ur.requests[searchByPrefixFileName]))
	result, err := ur.worker.Query(ctx, ur.requests[searchByPrefixFileName], pattern, ur.realmID, limit)

	if err != nil {
		ur.logger.ErrorContext(ctx, "Error executing query", slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, err
	}

	users := make([]model.UserSuggestion, 0, limit)
	for result.Next() {
		var u model.UserSuggestion
		if scanErr := result.Scan(&u.ID, &u.Login); scanErr != nil {
			ur.logger.ErrorContext(ctx, "Error scanning row",
				slog.String(consts.ErrorLoggerKey, scanErr.Error()))
			_ = result.Close()
			return nil, scanErr
		}
		users = append(users, u)
	}

	if err = result.Close(); err != nil {
		ur.logger.ErrorContext(ctx, "Error closing result",
			slog.String(consts.ErrorLoggerKey, err.Error()),
			slog.String(userPrefixKey, prefix))
		return nil, err
	}

	return users, nil
}

// escapeLikePattern escapes LIKE special characters (%, _, \) in the input string
// so they are treated as literals in a SQL LIKE expression.
func escapeLikePattern(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
}

func (ur *Repo) IDByName(ctx context.Context, login string) (model.UserID, error) {
	ur.logger.InfoContext(ctx, "About to execute query", slog.String("query_name", ur.requests[getUserByNameFileName]))
	result, err := ur.worker.Query(ctx, ur.requests[getUserByNameFileName], login, ur.realmID)

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
