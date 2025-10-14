package configs

import (
	"time"

	"github.com/dnonakolesax/viper"
)

const (
	servicePortKey                = "service.port"
	servicePortDefault            = 8800
	serviceAuthTimeoutKey         = "service.auth-timeout"
	serviceAuthTimeoutDefault     = 5 * time.Minute
	serviceBasePathKey            = "service.base-path"
	serviceBasePathDefault        = "/iam"
	serviceAllowedRedirectKey     = "service.allowed-redirect"
	serviceMetricsPortKey         = "service.metrics-port"
	serviceMetricsPortDefault     = 8801
	serviceGRPCPortKey            = "service.grpc-port"
	serviceGRPCPortDefault        = 8802
	serviceMetricsEndpointKey     = "service.metrics-endpoint"
	serviceMetricsEndpointDefault = "/metrics"
)

const (
	logLevelKey         = "service.log-level"
	logLevelDefault     = "info"
	logAddSourceKey     = "service.log-add-source"
	logAddSourceDefault = true
	logTimeoutKey       = "service.log-timeout"
	logTimeoutDefault   = 10 * time.Second
)

type ServiceConfig struct {
	Port            int
	AuthTimeout     time.Duration
	BasePath        string
	AllowedRedirect string
	MetricsPort     int
	GRPCPort        int
	MetricsEndpoint string
}

type LoggerConfig struct {
	LogLevel     string
	LogTimeout   time.Duration
	LogAddSource bool
}

func (sc *ServiceConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(servicePortKey, servicePortDefault)
	v.SetDefault(serviceAuthTimeoutKey, serviceAuthTimeoutDefault)
	v.SetDefault(serviceBasePathKey, serviceBasePathDefault)
	v.SetDefault(serviceAllowedRedirectKey, nil)
	v.SetDefault(serviceMetricsPortKey, serviceMetricsPortDefault)
	v.SetDefault(serviceGRPCPortKey, serviceGRPCPortDefault)
	v.SetDefault(serviceMetricsEndpointKey, serviceMetricsEndpointDefault)
}

func (sc *ServiceConfig) Load(v *viper.Viper) {
	sc.Port = v.GetInt(servicePortKey)
	sc.AuthTimeout = v.GetDuration(serviceAuthTimeoutKey)
	sc.BasePath = v.GetString(serviceBasePathKey)
	sc.AllowedRedirect = v.GetString(serviceAllowedRedirectKey)
	sc.MetricsPort = v.GetInt(serviceMetricsPortKey)
	sc.GRPCPort = v.GetInt(serviceGRPCPortKey)
	sc.MetricsEndpoint = v.GetString(serviceMetricsEndpointKey)
}

func (lc *LoggerConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(logLevelKey, logLevelDefault)
	v.SetDefault(logAddSourceKey, logAddSourceDefault)
	v.SetDefault(logTimeoutKey, logTimeoutDefault)
}

func (lc *LoggerConfig) Load(v *viper.Viper) {
	lc.LogLevel = v.GetString(logLevelKey)
	lc.LogAddSource = v.GetBool(logAddSourceKey)
	lc.LogTimeout = v.GetDuration(logTimeoutKey)
}
