package redis

import (
	"context"
	"strconv"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg configs.RedisConfig) (*redis.Client, error) {
	options := &redis.Options{
		Addr:     cfg.Address + ":" + strconv.Itoa(int(cfg.Port)),
		Password: cfg.Password,
		DB:       0,
	}

	client := redis.NewClient(options)

	err := client.Ping(context.TODO()).Err()

	if err != nil {
		return nil, err
	}

	return client, nil
}
