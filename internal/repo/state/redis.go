package state

import (
	"context"
	"time"

	"github.com/dnonakolesax/noted-auth/internal/errorvals"
	"github.com/redis/go-redis/v9"

	dbredis "github.com/dnonakolesax/noted-auth/internal/db/redis"
)

type RedisStateRepo struct {
	client *dbredis.RedisClient
}

func (rr *RedisStateRepo) SetState(state string, redirectURI string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), rr.client.Timeout)
	defer cancel()
	rsp := rr.client.Client.Set(ctx, state, redirectURI, time.Second*timeout)

	if rsp.Err() != nil {
		return rsp.Err()
	}

	return nil
}

func (rr *RedisStateRepo) GetState(state string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), rr.client.Timeout)
	defer cancel()
	val, err := rr.client.Client.Get(ctx, state).Result()

	if err == redis.Nil {
		return "", errorvals.ObjectNotFoundInRepoError
	} else if err != nil {
		return "", err
	}

	return val, nil
}

func NewRedisStateRepo(client *dbredis.RedisClient) *RedisStateRepo {
	return &RedisStateRepo{
		client: client,
	}
}
