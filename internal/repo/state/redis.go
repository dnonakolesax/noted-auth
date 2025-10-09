package state

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	dbredis "github.com/dnonakolesax/noted-auth/internal/db/redis"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
)

type RedisStateRepo struct {
	client *dbredis.Client
}

func NewRedisStateRepo(client *dbredis.Client) *RedisStateRepo {
	return &RedisStateRepo{
		client: client,
	}
}

func (rr *RedisStateRepo) SetState(state string, redirectURI string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), rr.client.Timeout)
	defer cancel()
	rsp := rr.client.Client.Set(ctx, state, redirectURI, timeout)

	if rsp.Err() != nil {
		return rsp.Err()
	}

	return nil
}

func (rr *RedisStateRepo) GetState(state string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rr.client.Timeout)
	defer cancel()
	val, err := rr.client.Client.Get(ctx, state).Result()

	if errors.Is(err, redis.Nil) {
		return "", errorvals.ErrObjectNotFoundInRepoError
	} else if err != nil {
		return "", err
	}

	return val, nil
}
