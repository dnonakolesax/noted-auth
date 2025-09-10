package repo

import (
	"context"
	"fmt"

	dbsql "github.com/dnonakolesax/noted-auth/internal/db/sql"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type UserRepo struct {
	worker *dbsql.PGXWorker
}

func (ur *UserRepo) GetUser(userId string) (user model.User, funcErr error) {
	result, err := ur.worker.Query(context.TODO(), ur.worker.Requests["get_user"], userId)

	defer func() {
		err := result.Close()
		if err != nil {
			funcErr = err
		}
	}()

	if err != nil {
		return model.User{}, err
	}

	if !result.Next() {
		return model.User{}, fmt.Errorf("not found")
	}
	err = result.Scan(&user.Login, &user.FirstName, &user.LastName) 
	if err != nil {
		return model.User{}, err
	}

	if result.Next() {
		return model.User{}, fmt.Errorf("too many rows")
	}
	return
}

func NewUserRepo(worker *dbsql.PGXWorker) (*UserRepo, error) {
	UserRequests, err := dbsql.LoadSQLRequests("./internal/repo/user/sql_requests")

	if err != nil {
		return nil, err
	}

	for key, value := range UserRequests {
		worker.Requests[key] = value
	}

	return &UserRepo{
		worker: worker,
	}, nil
}