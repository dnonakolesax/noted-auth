package state

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	dbredis "github.com/dnonakolesax/noted-auth/internal/db/redis"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
)

type RedisStateRepo struct {
	client *dbredis.Client
	logger *slog.Logger
}

func NewRedisStateRepo(client *dbredis.Client, logger *slog.Logger) *RedisStateRepo {
	return &RedisStateRepo{
		client: client,
		logger: logger,
	}
}

func (rr *RedisStateRepo) SetState(ctx context.Context, state string, redirectURI string, timeout time.Duration) error {
	rr.logger.DebugContext(ctx, "Setting state", "state", state, "redirectURI", redirectURI)
	dbctx, cancel := context.WithTimeout(ctx, rr.client.Timeout)
	defer cancel()
	rsp := rr.client.Client.Set(dbctx, state, redirectURI, timeout)

	if rsp.Err() != nil {
		rr.logger.ErrorContext(ctx, "Failed to set state", "error", rsp.Err())
		return rsp.Err()
	}
	rr.logger.DebugContext(ctx, "Set state success")

	return nil
}

func (rr *RedisStateRepo) GetState(ctx context.Context, state string) (string, error) {
	rr.logger.DebugContext(ctx, "Getting state", "state", state)
	dbctx, cancel := context.WithTimeout(ctx, rr.client.Timeout)
	defer cancel()
	val, err := rr.client.Client.Get(dbctx, state).Result()

	if errors.Is(err, redis.Nil) {
		rr.logger.WarnContext(ctx, "State not found", slog.String("error", err.Error()))
		return "", errorvals.ErrObjectNotFoundInRepoError
	} else if err != nil {
		rr.logger.ErrorContext(ctx, "Failed to get state", slog.String("error", err.Error()))
		return "", err
	}
	rr.logger.DebugContext(ctx, "Got state")

	return val, nil
}
