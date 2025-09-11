package configs

import (
	"github.com/spf13/viper"
	"time"
)

type HTTPRetryPolicyConfig struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	RetryOnStatus map[int]bool
}

func NewHTTPRetryPolicyConfig(v *viper.Viper) HTTPRetryPolicyConfig {
	retriesConfig := v.GetIntSlice("http.retries.on-status")
	retryOnStatus := make(map[int]bool, len(retriesConfig))

	for _, status := range retriesConfig {
		retryOnStatus[status] = true
	}

	return HTTPRetryPolicyConfig{
		MaxAttempts:   v.GetInt("http-client.retries.max-attempts"),
		BaseDelay:     v.GetDuration("http-client.retries.base-delay"),
		MaxDelay:      v.GetDuration("http-client.retries.max-delay"),
		RetryOnStatus: retryOnStatus,
	}
}

type HTTPClientConfig struct {
	DialTimeout     time.Duration
	RequestTimeout  time.Duration
	KeepAlive       time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
	RetryPolicy     HTTPRetryPolicyConfig
}

func NewHTTPClientConfig(v *viper.Viper, retryPolicy HTTPRetryPolicyConfig) HTTPClientConfig {
	return HTTPClientConfig{
		DialTimeout:     v.GetDuration("http-client.dial-timeout"),
		RequestTimeout:  v.GetDuration("http-client.request-timeout"),
		KeepAlive:       v.GetDuration("http-client.keep-alive"),
		MaxIdleConns:    v.GetInt("http-client.max-idle-conns"),
		IdleConnTimeout: v.GetDuration("http-client.idle-conn-timeout"),
		RetryPolicy:     retryPolicy,
	}
}
