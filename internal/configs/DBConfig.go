package configs

import (
	"time"

	"github.com/spf13/viper"
)

type RDBConfig struct {
	Address  string
	Port     uint
	DBName   string
	Login    string
	Password string

	ConnTimeout       time.Duration
	MinConns          int
	MaxConns          int
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	RequestTimeout    time.Duration
}

type RedisConfig struct {
	Address  string
	Port     uint
	Password string
}

func NewRDBConfig(v *viper.Viper) RDBConfig {
	return RDBConfig{
		Address:  v.GetString("postgres.address"),
		Port:     v.GetUint("postgres.port"),
		DBName:   v.GetString("postgres.database"),
		Login:    v.GetString("postgres.user"),
		Password: v.GetString("POSTGRES_PASSWORD"),

		ConnTimeout:       v.GetDuration("postgres.conn-timeout"),
		MinConns:          v.GetInt("postgres.min-conns"),
		MaxConns:          v.GetInt("postgres.max-conns"),
		MaxConnLifetime:   v.GetDuration("postgres.max-conn-lifetime"),
		MaxConnIdleTime:   v.GetDuration("postgres.max-conn-idle-time"),
		HealthCheckPeriod: v.GetDuration("postgres.healthcheck-period"),
		RequestTimeout:    v.GetDuration("postgres.request-timeout"),
	}
}

func NewRedisConfig(v *viper.Viper) RedisConfig {
	return RedisConfig{
		Address:  v.GetString("redis.address"),
		Port:     v.GetUint("redis.port"),
		Password: v.GetString("REDIS_PASSWORD"),
	}
}
