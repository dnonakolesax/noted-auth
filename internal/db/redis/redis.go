package redis

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/consts"
)

type Client struct {
	Client  *redis.Client
	Timeout time.Duration
	logger  *slog.Logger
}

const addressLoggerKey = "address"

func NewClient(cfg configs.RedisConfig, logger *slog.Logger) (*Client, error) {
	options := &redis.Options{
		Addr:     cfg.Address + ":" + strconv.Itoa(cfg.Port),
		Password: cfg.Password,
		DB:       0,
	}

	logger.Info("Starting new redis client", slog.String(addressLoggerKey, options.Addr))
	client := redis.NewClient(options)
	logger.Info("Redis client started", slog.String(addressLoggerKey, options.Addr))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()
	logger.Info("Trying to ping redis client", slog.String(addressLoggerKey, options.Addr))
	err := client.Ping(ctx).Err()

	if err != nil {
		logger.Error("Error while pinging redis client", slog.String(addressLoggerKey, options.Addr),
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return nil, err
	}
	logger.Info("Redis client ping successfull", slog.String(addressLoggerKey, options.Addr))

	return &Client{
		Client:  client,
		Timeout: cfg.RequestTimeout,
		logger:  logger,
	}, nil
}
