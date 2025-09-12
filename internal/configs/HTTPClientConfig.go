package configs

import (
	"github.com/dnonakolesax/viper"
	"time"
)

const HTTPC_RETRY_MAX_ATTEMPTS_KEY = "http-client.retries.max-attempts"
const HTTPC_RETRY_BASE_DELAY_KEY = "http-client.retries.base-delay"
const HTTPC_RETRY_MAX_DELAY_KEY = "http-client.retries.max-delay"
const HTTPC_RETRY_ON_STATUS_KEY = "http-client.retries.on-status"
const HTTPC_DIAL_TIMEOUT_KEY = "http-client.dial-timeout"
const HTTPC_REQUEST_TIMEOUT_KEY = "http-client.request-timeout"
const HTTPC_KEEP_ALIVE_KEY = "http-client.keep-alive"
const HTTPC_MAX_IDLE_CONNS_KEY = "http-client.max-idle-conns"
const HTTPC_IDLE_CONN_TIMEOUT_KEY = "http-client.idle-conn-timeout"

type HTTPRetryPolicyConfig struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	RetryOnStatus map[int]bool
}

func (hc *HTTPRetryPolicyConfig) Load(v *viper.Viper) {
	retriesConfig := v.GetIntSlice(HTTPC_RETRY_ON_STATUS_KEY)
	retryOnStatus := make(map[int]bool, len(retriesConfig))

	for _, status := range retriesConfig {
		retryOnStatus[status] = true
	}

	hc.MaxAttempts = v.GetInt(HTTPC_RETRY_MAX_ATTEMPTS_KEY)
	hc.BaseDelay = v.GetDuration(HTTPC_RETRY_BASE_DELAY_KEY)
	hc.MaxDelay = v.GetDuration(HTTPC_RETRY_MAX_DELAY_KEY)
	hc.RetryOnStatus = retryOnStatus
}

func (hc *HTTPRetryPolicyConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(HTTPC_RETRY_MAX_ATTEMPTS_KEY, 3)
	v.SetDefault(HTTPC_RETRY_BASE_DELAY_KEY, 200*time.Millisecond)
	v.SetDefault(HTTPC_RETRY_MAX_DELAY_KEY, 3*time.Second)
	v.SetDefault(HTTPC_RETRY_ON_STATUS_KEY, []int{429, 502, 503, 504})
}

type HTTPClientConfig struct {
	DialTimeout     time.Duration
	RequestTimeout  time.Duration
	KeepAlive       time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
	RetryPolicy     HTTPRetryPolicyConfig
}

func (hc *HTTPClientConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(HTTPC_DIAL_TIMEOUT_KEY, 5*time.Second)
	v.SetDefault(HTTPC_REQUEST_TIMEOUT_KEY, 30*time.Second)
	v.SetDefault(HTTPC_KEEP_ALIVE_KEY, 30*time.Second)
	v.SetDefault(HTTPC_MAX_IDLE_CONNS_KEY, 100)
	v.SetDefault(HTTPC_IDLE_CONN_TIMEOUT_KEY, 90*time.Second)
	hc.RetryPolicy.SetDefaults(v)
}

func (hc *HTTPClientConfig) Load(v *viper.Viper) {
	hc.DialTimeout = v.GetDuration(HTTPC_DIAL_TIMEOUT_KEY)
	hc.RequestTimeout = v.GetDuration(HTTPC_REQUEST_TIMEOUT_KEY)
	hc.KeepAlive = v.GetDuration(HTTPC_KEEP_ALIVE_KEY)
	hc.MaxIdleConns = v.GetInt(HTTPC_MAX_IDLE_CONNS_KEY)
	hc.IdleConnTimeout = v.GetDuration(HTTPC_IDLE_CONN_TIMEOUT_KEY)
	hc.RetryPolicy.Load(v)
}
