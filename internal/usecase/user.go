package usecase

import (
	"context"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/model"
)

type repo interface {
	GetUser(ctx context.Context, userID string) (model.User, error)
	IDByName(ctx context.Context, login string) (model.UserID, error)
	SearchByPrefix(ctx context.Context, prefix string, limit int) ([]model.UserSuggestion, error)
}

type UserUsecase struct {
	userRepo repo
	logger   *slog.Logger
}

func NewUserUsecase(userRepo repo, logger *slog.Logger) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (uu *UserUsecase) Get(ctx context.Context, userID string) (model.User, error) {
	user, err := uu.userRepo.GetUser(ctx, userID)

	if err != nil {
		uu.logger.ErrorContext(ctx, "Error getting user",
			slog.String(consts.ErrorLoggerKey, err.Error()), slog.String("ID", userID))
		return model.User{}, err
	}

	return user, nil
}

func (uu *UserUsecase) SearchByPrefix(ctx context.Context, prefix string) (model.UserSuggestions, error) {
	users, err := uu.userRepo.SearchByPrefix(ctx, prefix, consts.UserSearchByPrefixLimit)

	if err != nil {
		uu.logger.ErrorContext(ctx, "Error searching users by prefix",
			slog.String(consts.ErrorLoggerKey, err.Error()), slog.String("PREFIX", prefix))
		return model.UserSuggestions{}, err
	}

	return model.UserSuggestions{Items: users}, nil
}

func (uu *UserUsecase) GetByUsername(ctx context.Context, username string) (model.UserID, error) {
	user, err := uu.userRepo.IDByName(ctx, username)

	if err != nil {
		uu.logger.ErrorContext(ctx, "Error getting user",
			slog.String(consts.ErrorLoggerKey, err.Error()), slog.String("LOGIN", username))
		return model.UserID{}, err
	}

	return user, nil
}
