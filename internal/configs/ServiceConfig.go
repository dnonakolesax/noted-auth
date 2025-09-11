package configs

import (
	"time"

	"github.com/spf13/viper"
)

type ServiceConfig struct {
	Port            uint
	AuthTimeout     time.Duration
	BasePath        string
	AllowedRedirect string
	MetricsPort     uint
	GRPCPort        uint
	LogLevel        string
	LogAddSource    bool
	MetricsEndpoint string
}

func NewServiceConfig(v *viper.Viper) ServiceConfig {
	return ServiceConfig{
		Port:            v.GetUint("service.port"),
		AuthTimeout:     v.GetDuration("service.auth-timeout"),
		BasePath:        v.GetString("service.base-path"),
		AllowedRedirect: v.GetString("service.allowed-redirect"),
		MetricsPort:     v.GetUint("service.metrics-port"),
		GRPCPort:        v.GetUint("service.grpc-port"),
		LogLevel:        v.GetString("service.log-level"),
		LogAddSource:    v.GetBool("service.log-add-source"),
		MetricsEndpoint: v.GetString("service.metrics-endpoint"),
	}
}
