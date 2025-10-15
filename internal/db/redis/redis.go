package redis

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/errorvals"
)

type Client struct {
	Client  *redis.Client
	Timeout time.Duration
	logger  *slog.Logger
	Alive   *atomic.Bool
}

const addressLoggerKey = "address"

func NewClient(cfg configs.RedisConfig, alive *atomic.Bool, logger *slog.Logger) (*Client, error) {
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

	alive.Store(true)
	return &Client{
		Client:  client,
		Timeout: cfg.RequestTimeout,
		logger:  logger,
		Alive:   alive,
	}, nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	rctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	val, err := c.Client.Get(rctx, key).Result()

	if errors.Is(err, redis.Nil) {
		c.logger.WarnContext(ctx, "State not found", slog.String(consts.ErrorLoggerKey, err.Error()))
		return "", errorvals.ErrObjectNotFoundInRepoError
	} else if err != nil {
		c.Alive.Store(false)
		c.logger.ErrorContext(ctx, "Error getting value from redis", slog.String(consts.ErrorLoggerKey, err.Error()))
		return "", err
	}

	return val, nil
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	rctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()
	err := c.Client.Set(rctx, key, value, ttl).Err()

	if err != nil {
		c.Alive.Store(false)
		c.logger.ErrorContext(ctx, "Failed to set state to redis", slog.String(consts.ErrorLoggerKey, err.Error()))
		return err
	}
	return nil
}
