package configs

import (
	"time"

	"github.com/spf13/viper"
)

type KeycloakConfig struct {
	ClientId              string
	ClientSecret          string
	RedirectURI           string
	RealmAddress          string
	StateLength           uint
	TokenTimeout          time.Duration
	RealmId               string
	PostLogoutRedirectURI string
	AuthEndpoint          string
	TokenEndpoint         string
	LogoutEndpoint        string
}

func NewKeycloakConfig(v *viper.Viper) KeycloakConfig {
	return KeycloakConfig{
		ClientId:              v.GetString("realm.client-id"),
		ClientSecret:          v.GetString("REALM_CLIENT_SECRET"),
		RedirectURI:           v.GetString("realm.redirect-url"),
		RealmAddress:          v.GetString("realm.base-url"),
		StateLength:           v.GetUint("realm.state-length"),
		TokenTimeout:          v.GetDuration("realm.te-timeout"),
		RealmId:               v.GetString("realm.id"),
		PostLogoutRedirectURI: v.GetString("realm.post-logout-redirect-uri"),
		AuthEndpoint:          v.GetString("realm.auth-endpoint"),
		TokenEndpoint:         v.GetString("realm.token-endpoint"),
		LogoutEndpoint:        v.GetString("realm.logout-endpoint"),
	}
}
