package configs

import (
	"time"

	"github.com/spf13/viper"
)

type KeycloakConfig struct {
	ClientId string
	ClientSecret string
	RedirectURI string
	RealmAddress string
	StateLength uint
	TokenTimeout time.Duration
}

func NewKeycloakConfig(v *viper.Viper) KeycloakConfig {
	return KeycloakConfig{
		ClientId: v.GetString("realm.client-id"),
		ClientSecret: v.GetString("realm.client-secret"),
		RedirectURI: v.GetString("realm.redirect-url"),
		RealmAddress: v.GetString("realm.base-url"),
		StateLength: v.GetUint("realm.state-length"),
		TokenTimeout: v.GetDuration("realm.te-timeout"),
	}
}
