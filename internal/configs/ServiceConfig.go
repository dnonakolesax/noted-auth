package configs

import (
	"time"

	"github.com/dnonakolesax/viper"
)

const SERVICE_PORT_KEY = "service.port"
const SERVICE_AUTH_TIMEOUT_KEY = "service.auth-timeout"
const SERVICE_BASE_PATH_KEY = "service.base-path"
const SERVICE_ALLOWED_REDIRECT_KEY = "service.allowed-redirect"
const SERVICE_METRICS_PORT_KEY = "service.metrics-port"
const SERVICE_GRPC_PORT_KEY = "service.grpc-port"
const SERVICE_METRICS_ENDPOINT_KEY = "service.metrics-endpoint"

const LOG_LEVEL_KEY = "service.log-level"
const LOG_ADD_SOURCE_KEY = "service.log-add-source"

type ServiceConfig struct {
	Port            uint
	AuthTimeout     time.Duration
	BasePath        string
	AllowedRedirect string
	MetricsPort     uint
	GRPCPort        uint
	MetricsEndpoint string
}

type LoggerConfig struct {
	LogLevel     string
	LogAddSource bool
}

func (sc *ServiceConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(SERVICE_PORT_KEY, 8800)
	v.SetDefault(SERVICE_AUTH_TIMEOUT_KEY, 5*time.Minute)
	v.SetDefault(SERVICE_BASE_PATH_KEY, "/iam")
	v.SetDefault(SERVICE_ALLOWED_REDIRECT_KEY, nil)
	v.SetDefault(SERVICE_METRICS_PORT_KEY, 8801)
	v.SetDefault(SERVICE_GRPC_PORT_KEY, 8802)
	v.SetDefault(SERVICE_METRICS_ENDPOINT_KEY, "/metrics")
}

func (sc *ServiceConfig) Load(v *viper.Viper) {
	sc.Port = v.GetUint(SERVICE_PORT_KEY)
	sc.AuthTimeout = v.GetDuration(SERVICE_AUTH_TIMEOUT_KEY)
	sc.BasePath = v.GetString(SERVICE_BASE_PATH_KEY)
	sc.AllowedRedirect = v.GetString(SERVICE_ALLOWED_REDIRECT_KEY)
	sc.MetricsPort = v.GetUint(SERVICE_METRICS_PORT_KEY)
	sc.GRPCPort = v.GetUint(SERVICE_GRPC_PORT_KEY)
	sc.MetricsEndpoint = v.GetString(SERVICE_METRICS_ENDPOINT_KEY)
}

func (lc *LoggerConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(LOG_LEVEL_KEY, "info")
	v.SetDefault(LOG_ADD_SOURCE_KEY, true)
}

func (lc *LoggerConfig) Load(v *viper.Viper) {
	lc.LogLevel = v.GetString(LOG_LEVEL_KEY)
	lc.LogAddSource = v.GetBool(LOG_ADD_SOURCE_KEY)
}
