package configs

import (
	"time"

	"github.com/dnonakolesax/viper"
)

const REALM_CLIENT_ID_KEY = "realm.client-id"
const REALM_CLIENT_SERCRET_KEY = "realm_client_secret"
const REALM_REDIRECT_URL_KEY = "realm.redirect-url"
const REALM_ADDRESS_KEY = "realm.base-url"
const REALM_STATE_LENGTH_KEY = "realm.state-length"
const REALM_TOKEN_TIMEOUT_KEY = "realm.te-timeout"
const REALM_ID_KEY = "realm.id"
const REALM_POST_LOGOUT_REDIRECT_URI_KEY = "realm.post-logout-redirect-uri"
const REALM_AUTH_ENDPOINT_KEY = "realm.auth-endpoint"
const REALM_TOKEN_ENDPOINT_KEY = "realm.token-endpoint"
const REALM_LOGOUT_ENDPOINT_KEY = "realm.logout-endpoint"

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

func (kc *KeycloakConfig) Load(v *viper.Viper) {
	kc.ClientId = v.GetString(REALM_CLIENT_ID_KEY)
	kc.ClientSecret = v.GetString(REALM_CLIENT_SERCRET_KEY)
	kc.RedirectURI = v.GetString(REALM_REDIRECT_URL_KEY)
	kc.RealmAddress = v.GetString(REALM_ADDRESS_KEY)
	kc.StateLength = v.GetUint(REALM_STATE_LENGTH_KEY)
	kc.TokenTimeout = v.GetDuration(REALM_TOKEN_TIMEOUT_KEY)
	kc.RealmId = v.GetString(REALM_ID_KEY)
	kc.PostLogoutRedirectURI = v.GetString(REALM_POST_LOGOUT_REDIRECT_URI_KEY)
	kc.AuthEndpoint = v.GetString(REALM_AUTH_ENDPOINT_KEY)
	kc.TokenEndpoint = v.GetString(REALM_TOKEN_ENDPOINT_KEY)
	kc.LogoutEndpoint = v.GetString(REALM_LOGOUT_ENDPOINT_KEY)
}

func (kc *KeycloakConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(REALM_CLIENT_ID_KEY, "webpage")
	v.SetDefault(REALM_CLIENT_SERCRET_KEY, nil)
	v.SetDefault(REALM_REDIRECT_URL_KEY, nil)
	v.SetDefault(REALM_ADDRESS_KEY, nil)
	v.SetDefault(REALM_STATE_LENGTH_KEY, 32)
	v.SetDefault(REALM_TOKEN_TIMEOUT_KEY, 10*time.Second)
	v.SetDefault(REALM_ID_KEY, nil)
	v.SetDefault(REALM_POST_LOGOUT_REDIRECT_URI_KEY, nil)
	v.SetDefault(REALM_AUTH_ENDPOINT_KEY, "/auth")
	v.SetDefault(REALM_TOKEN_ENDPOINT_KEY, "/token")
	v.SetDefault(REALM_LOGOUT_ENDPOINT_KEY, "/logout")
}
