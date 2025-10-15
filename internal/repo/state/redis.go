package state

import (
	"context"
	"log/slog"
	"time"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	dbredis "github.com/dnonakolesax/noted-auth/internal/db/redis"
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
		rr.logger.ErrorContext(ctx, "Failed to set state", slog.String(consts.ErrorLoggerKey, rsp.Err().Error()))
		return rsp.Err()
	}
	rr.logger.DebugContext(ctx, "Set state success")

	return nil
}

func (rr *RedisStateRepo) GetState(ctx context.Context, state string) (string, error) {
	rr.logger.DebugContext(ctx, "Getting state", "state", state)
	val, err := rr.client.Get(ctx, state)

	if err != nil {
		rr.logger.ErrorContext(ctx, "Failed to get state", slog.String(consts.ErrorLoggerKey, err.Error()))
		return "", err
	}

	rr.logger.DebugContext(ctx, "Got state")

	return val, nil
}
