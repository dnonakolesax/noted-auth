package configs

import (
	"time"

	"github.com/dnonakolesax/viper"
)

const POSTGRES_ADDRESS_KEY = "postgres.address"
const POSTGRES_PORT_KEY = "postgres.port"
const POSTGRES_DB_NAME_KEY = "postgres.database"
const POSTGRES_LOGIN_KEY = "postgres.user"
const POSTGRES_PASSWORD_KEY = "postgres_password"

const POSTGRES_CONN_TIMEOUT_KEY = "postgres.conn-timeout"
const POSTGRES_MIN_CONNS_KEY = "postgres.min-conns"
const POSTGRES_MAX_CONNS_KEY = "postgres.max-conns"
const POSTGRES_MAX_CONN_LIFETIME_KEY = "postgres.max-conn-lifetime"
const POSTGRES_MAX_CONN_IDLE_TIME_KEY = "postgres.max-conn-idle-time"
const POSTGRES_HEALTHCHECK_PERIOD_KEY = "postgres.healthcheck-period"
const POSTGRES_REQUEST_TIMEOUT_KEY = "postgres.request-timeout"

const REDIS_ADDRESS_KEY = "redis.address"
const REDIS_PORT_KEY = "redis.port"
const REDIS_PASSWORD_KEY = "redis_password"

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

func (rc *RDBConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(POSTGRES_ADDRESS_KEY, "postgres")
	v.SetDefault(POSTGRES_PORT_KEY, 5432)
	v.SetDefault(POSTGRES_DB_NAME_KEY, "keycloak")
	v.SetDefault(POSTGRES_LOGIN_KEY, nil)
	v.SetDefault(POSTGRES_PASSWORD_KEY, nil)

	v.SetDefault(POSTGRES_CONN_TIMEOUT_KEY, 10*time.Second)
	v.SetDefault(POSTGRES_MIN_CONNS_KEY, 5)
	v.SetDefault(POSTGRES_MAX_CONNS_KEY, 20)
	v.SetDefault(POSTGRES_MAX_CONN_LIFETIME_KEY, time.Hour)
	v.SetDefault(POSTGRES_MAX_CONN_IDLE_TIME_KEY, 30*time.Minute)
	v.SetDefault(POSTGRES_HEALTHCHECK_PERIOD_KEY, time.Minute)
	v.SetDefault(HTTPC_REQUEST_TIMEOUT_KEY, 30*time.Second)
}

func (rc *RDBConfig) Load(v *viper.Viper) {
	rc.Address = v.GetString(POSTGRES_ADDRESS_KEY)
	rc.Port = v.GetUint(POSTGRES_PORT_KEY)
	rc.DBName = v.GetString(POSTGRES_DB_NAME_KEY)
	rc.Login = v.GetString(POSTGRES_LOGIN_KEY)
	rc.Password = v.GetString(POSTGRES_PASSWORD_KEY)

	rc.ConnTimeout = v.GetDuration(POSTGRES_CONN_TIMEOUT_KEY)
	rc.MinConns = v.GetInt(POSTGRES_MIN_CONNS_KEY)
	rc.MaxConns = v.GetInt(POSTGRES_MAX_CONNS_KEY)
	rc.MaxConnLifetime = v.GetDuration(POSTGRES_MAX_CONN_LIFETIME_KEY)
	rc.MaxConnIdleTime = v.GetDuration(POSTGRES_MAX_CONN_IDLE_TIME_KEY)
	rc.HealthCheckPeriod = v.GetDuration(POSTGRES_HEALTHCHECK_PERIOD_KEY)
	rc.RequestTimeout = v.GetDuration(HTTPC_REQUEST_TIMEOUT_KEY)
}

func (rc *RedisConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(REDIS_ADDRESS_KEY, "redis")
	v.SetDefault(REDIS_PORT_KEY, 6379)
	v.SetDefault(REDIS_PASSWORD_KEY, nil)
}

func (rc *RedisConfig) Load(v *viper.Viper) {
	rc.Address = v.GetString(REDIS_ADDRESS_KEY)
	rc.Port = v.GetUint(REDIS_PORT_KEY)
	rc.Password = v.GetString(REDIS_PASSWORD_KEY)
}
