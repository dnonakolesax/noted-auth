package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dnonakolesax/noted-auth/internal/configs"
)

type Client struct {
	Client  *redis.Client
	Timeout time.Duration
}

func NewClient(cfg configs.RedisConfig) (*Client, error) {
	options := &redis.Options{
		Addr:     cfg.Address + ":" + strconv.Itoa(cfg.Port),
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

	return &Client{
		Client:  client,
		Timeout: cfg.RequestTimeout,
	}, nil
}
