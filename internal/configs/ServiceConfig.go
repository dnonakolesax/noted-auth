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
	logDirKey             = "service.log-dir"
	logDirDefault         = "/var/log/noted-auth"
	logLevelKey           = "service.log-level"
	logLevelDefault       = "info"
	logAddSourceKey       = "service.log-add-source"
	logAddSourceDefault   = true
	logTimeoutKey         = "service.log-timeout"
	logTimeoutDefault     = 10 * time.Second
	logMaxFileSizeKey     = "service.log-max-file-size"
	logMaxFileSizeDefault = 100
	logMaxBackupsKey      = "service.log-max-backups"
	logMaxBackupsDefault  = 3
	logMaxAgeKey          = "service.log-max-age"
	logMaxAgeDefault      = 28
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
	LogDir         string
	LogLevel       string
	LogTimeout     time.Duration
	LogAddSource   bool
	LogMaxFileSize int
	LogMaxBackups  int
	LogMaxAge      int
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
	v.SetDefault(logDirKey, logDirDefault)
	v.SetDefault(logLevelKey, logLevelDefault)
	v.SetDefault(logAddSourceKey, logAddSourceDefault)
	v.SetDefault(logTimeoutKey, logTimeoutDefault)
	v.SetDefault(logMaxFileSizeKey, logMaxFileSizeDefault)
	v.SetDefault(logMaxBackupsKey, logMaxBackupsDefault)
	v.SetDefault(logMaxAgeKey, logMaxAgeDefault)
}

func (lc *LoggerConfig) Load(v *viper.Viper) {
	lc.LogDir = v.GetString(logDirKey)
	lc.LogLevel = v.GetString(logLevelKey)
	lc.LogAddSource = v.GetBool(logAddSourceKey)
	lc.LogTimeout = v.GetDuration(logTimeoutKey)
	lc.LogMaxFileSize = v.GetInt(logMaxFileSizeKey)
	lc.LogMaxBackups = v.GetInt(logMaxBackupsKey)
	lc.LogMaxAge = v.GetInt(logMaxAgeKey)
}
