package usecase

import "github.com/dnonakolesax/noted-auth/internal/model"

type UserRepo interface {
	GetUser(userID string) (model.User, error)
}

type UserUsecase struct {
	userRepo UserRepo
}

func NewUserUsecase(userRepo UserRepo) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (uu *UserUsecase) Get(userID string) (model.User, error) {
	user, err := uu.userRepo.GetUser(userID)

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
