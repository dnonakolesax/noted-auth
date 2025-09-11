package state

import (
	"context"
	"time"

	"github.com/dnonakolesax/noted-auth/internal/errorvals"
	"github.com/redis/go-redis/v9"
)

type RedisStateRepo struct {
	client *redis.Client
}

func (rr *RedisStateRepo) SetState(state string, redirectURI string, timeout time.Duration) error {
	rsp := rr.client.Set(context.TODO(), state, redirectURI, time.Second*timeout)

	if rsp.Err() != nil {
		return rsp.Err()
	}

	return nil
}

func (rr *RedisStateRepo) GetState(state string) (string, error) {
	val, err := rr.client.Get(context.TODO(), state).Result()

	if err == redis.Nil {
		return "", errorvals.ObjectNotFoundInRepoError
	} else if err != nil {
		return "", err
	}

	return val, nil
}

func NewRedisStateRepo(client *redis.Client) *RedisStateRepo {
	return &RedisStateRepo{
		client: client,
	}
}
