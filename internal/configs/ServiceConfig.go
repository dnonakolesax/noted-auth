package configs

import (
	"time"

	"github.com/spf13/viper"
)

type ServiceConfig struct {
	Port 		    uint
	AuthTimeout     time.Duration
	BasePath	    string
	AllowedRedirect string
}

func NewServiceConfig(v *viper.Viper) ServiceConfig {
	return ServiceConfig{
		Port: viper.GetUint("app.port"),
		AuthTimeout: viper.GetDuration("app.auth-timeout"),
		BasePath: viper.GetString("app.base-path"),
		AllowedRedirect: viper.GetString("app.allowed-redirect"),
	}
}
