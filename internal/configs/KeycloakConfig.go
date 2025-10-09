package configs

import (
	"time"

	"github.com/dnonakolesax/viper"
)

const (
	realmClientIDKey           = "realm.client-id"
	realmClientIDDefault       = "webpage"
	realmClientSecretKey       = "realm_client_secret"
	realmRedirectURLKey        = "realm.redirect-url"
	realmAddressKey            = "realm.base-url"
	realmInterAddressKey       = "realm.inter-url"
	realmStateLengthKey        = "realm.state-length"
	realmDefaultStateLength    = 32
	realmTokenTimeoutKey       = "realm.te-timeout"
	realmTokenTimeoutDefault   = 10 * time.Second
	realmIDKey                 = "realm.id"
	realmPostLogoutURIKey      = "realm.post-logout-redirect-uri"
	realmAuthEndpointKey       = "realm.auth-endpoint"
	realmAuthEndpointDefault   = "/auth"
	realmTokenEndpointKey      = "realm.token-endpoint"
	realmTokenEndpointDefault  = "/token"
	realmLogoutEndpointKey     = "realm.logout-endpoint"
	realmLogoutEndpointDefault = "/logout"
)

type KeycloakConfig struct {
	ClientID              string
	ClientSecret          string
	RedirectURI           string
	RealmAddress          string
	InterRealmAddress     string
	StateLength           uint
	TokenTimeout          time.Duration
	RealmID               string
	PostLogoutRedirectURI string
	AuthEndpoint          string
	TokenEndpoint         string
	LogoutEndpoint        string
}

func (kc *KeycloakConfig) Load(v *viper.Viper) {
	kc.ClientID = v.GetString(realmClientIDKey)
	kc.ClientSecret = v.GetString(realmClientSecretKey)
	kc.RedirectURI = v.GetString(realmRedirectURLKey)
	kc.RealmAddress = v.GetString(realmAddressKey)
	v.SetDefault(realmInterAddressKey, kc.RealmAddress)
	kc.InterRealmAddress = v.GetString(realmInterAddressKey)
	kc.StateLength = v.GetUint(realmStateLengthKey)
	kc.TokenTimeout = v.GetDuration(realmTokenTimeoutKey)
	kc.RealmID = v.GetString(realmIDKey)
	kc.PostLogoutRedirectURI = v.GetString(realmPostLogoutURIKey)
	kc.AuthEndpoint = v.GetString(realmAuthEndpointKey)
	kc.TokenEndpoint = v.GetString(realmTokenEndpointKey)
	kc.LogoutEndpoint = v.GetString(realmLogoutEndpointKey)
}

func (kc *KeycloakConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(realmClientIDKey, realmClientIDDefault)
	v.SetDefault(realmClientSecretKey, nil)
	v.SetDefault(realmRedirectURLKey, nil)
	v.SetDefault(realmAddressKey, nil)
	v.SetDefault(realmInterAddressKey, nil)
	v.SetDefault(realmStateLengthKey, realmDefaultStateLength)
	v.SetDefault(realmTokenTimeoutKey, realmTokenTimeoutDefault)
	v.SetDefault(realmIDKey, nil)
	v.SetDefault(realmPostLogoutURIKey, nil)
	v.SetDefault(realmAuthEndpointKey, realmAuthEndpointDefault)
	v.SetDefault(realmTokenEndpointKey, realmTokenEndpointDefault)
	v.SetDefault(realmLogoutEndpointKey, realmLogoutEndpointDefault)
}
