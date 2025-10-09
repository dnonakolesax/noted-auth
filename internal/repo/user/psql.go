package repo

import (
	"context"
	"errors"

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
}

func NewUserRepo(worker *dbsql.PGXWorker, realmID string, requestsPath string) (*UserRepo, error) {
	userRequests, err := dbsql.LoadSQLRequests(requestsPath + thisDomainName)

	if err != nil {
		return nil, err
	}

	for key, value := range userRequests {
		worker.Requests[key] = value
	}

	return &UserRepo{
		worker:  worker,
		realmID: realmID,
	}, nil
}

func (ur *UserRepo) GetUser(userID string) (model.User, error) {
	result, err := ur.worker.Query(context.TODO(), ur.worker.Requests[getUserFileName], userID, ur.realmID)

	if err != nil {
		return model.User{}, err
	}

	if !result.Next() {
		return model.User{}, errors.New("not found")
	}
	var user model.User
	err = result.Scan(&user.Login, &user.FirstName, &user.LastName)
	if err != nil {
		return model.User{}, err
	}

	if result.Next() {
		return model.User{}, errors.New("too many rows")
	}

	err = result.Close()
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
