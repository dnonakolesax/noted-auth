package state

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/muesli/cache2go"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
)

type InMemStateRepo struct {
	client *cache2go.CacheTable
	logger *slog.Logger
}

func NewInMemStateRepo(logger *slog.Logger) *InMemStateRepo {
	return &InMemStateRepo{
		logger: logger,
	}
}

func (sr *InMemStateRepo) SetState(ctx context.Context, state string, redirectURI string, timeout time.Duration) error {
	sr.logger.DebugContext(ctx, "Adding state to in-memory cache")
	sr.client.Add(state, timeout, redirectURI)

	return nil
}

func (sr *InMemStateRepo) GetState(ctx context.Context, state string) (string, error) {
	sr.logger.DebugContext(ctx, "Getting state from in-memory cache")
	val, err := sr.client.Value(state)

	if err != nil {
		if errors.Is(err, cache2go.ErrKeyNotFound) {
			sr.logger.WarnContext(ctx, "Key not found in in-memory cache")
			return "", errorvals.ErrObjectNotFoundInRepoError
		}
		sr.logger.ErrorContext(ctx, "Error getting state from in-memory cache",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return "", err
	}
	sr.logger.DebugContext(ctx, "Got state from in-memory cache")

	stringData, ok := val.Data().(string)

	if !ok {
		sr.logger.ErrorContext(ctx, "Failed to cast in-memory cache data to string")
		return "", errors.New("failed to cast data to string")
	}

	return stringData, nil
}
