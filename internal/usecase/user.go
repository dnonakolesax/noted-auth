package usecase

import "github.com/dnonakolesax/noted-auth/internal/model"

type UserRepo interface {
	GetUser(userId string) (model.User, error)
}

type UserUsecase struct {
	userRepo UserRepo
}

func (uu *UserUsecase) Get(userId string) (model.User, error) {
	user, err := uu.userRepo.GetUser(userId)

	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func NewUserUsecase(userRepo UserRepo) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}
