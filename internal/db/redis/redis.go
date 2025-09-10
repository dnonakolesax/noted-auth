package redis

import (
	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/redis/go-redis/v9"
)

func NewClient(cfg configs.RedisConfig) *redis.Client {
	options := &redis.Options{
		Addr: cfg.Address + ":" + string(cfg.Port),
		Password: cfg.Password,
		DB: 0,
	}

	client := redis.NewClient(options)

	return client
}
