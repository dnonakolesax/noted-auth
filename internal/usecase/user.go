package usecase

import (
	"context"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/model"
)

type UserRepo interface {
	GetUser(ctx context.Context, userID string) (model.User, error)
}

type UserUsecase struct {
	userRepo UserRepo
	logger   *slog.Logger
}

func NewUserUsecase(userRepo UserRepo, logger *slog.Logger) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (uu *UserUsecase) Get(ctx context.Context, userID string) (model.User, error) {
	user, err := uu.userRepo.GetUser(ctx, userID)

	if err != nil {
		uu.logger.ErrorContext(ctx, "Error getting user", slog.String("error", err.Error()))
		return model.User{}, err
	}

	return user, nil
}
