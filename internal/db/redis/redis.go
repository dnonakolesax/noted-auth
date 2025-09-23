package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
	Timeout time.Duration
}

func NewClient(cfg configs.RedisConfig) (*RedisClient, error) {
	options := &redis.Options{
		Addr:     cfg.Address + ":" + strconv.Itoa(int(cfg.Port)),
		Password: cfg.Password,
		DB:       0,
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()
	err := client.Ping(ctx).Err()

	if err != nil {
		return nil, err
	}

	return &RedisClient{
		Client: client,
		Timeout: cfg.RequestTimeout,
	}, nil
}
