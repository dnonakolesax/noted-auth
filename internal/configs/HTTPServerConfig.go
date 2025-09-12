package configs

import (
	"time"

	"github.com/dnonakolesax/viper"
)

const HTTPS_READ_TIMEOUT_KEY = "http-server.read-timeout"
const HTTPS_WRITE_TIMEOUT_KEY = "http-server.write-timeout"
const HTTPS_IDLE_TIMEOUT_KEY = "http-server.idle-timeout"
const HTTPS_MAX_IDLE_WORKER_DURATION_KEY = "http-server.max-idle-worker-duration"
const HTTPS_MAX_REQ_BODY_SIZE_KEY = "http-server.max-req-body-size"
const HTTPS_READ_BUFFER_SIZE_KEY = "http-server.read-buffer-size"
const HTTPS_WRITE_BUFFER_SIZE_KEY = "http-server.write-buffer-size"
const HTTPS_CONCURRENCY_KEY = "http-server.concurrency"
const HTTPS_MAX_CONNS_PER_IP_KEY = "http-server.max-conns-per-ip"
const HTTPS_MAX_REQUESTS_PER_CONN_KEY = "http-server.max-requests-per-conn"
const HTTPS_TCP_KEEPALIVE_PERIOD_KEY = "http-server.tcp-keepalive-period"

type HTTPServerConfig struct {
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
	MaxIdleWorkerDuration time.Duration
	MaxReqBodySize        int
	ReadBufferSize        int
	WriteBufferSize       int
	Concurrency           int
	MaxConnsPerIP         int
	MaxRequestsPerConn    int
	TCPKeepAlivePeriod    time.Duration
}

func (hc *HTTPServerConfig) Load(v *viper.Viper) {
	hc.ReadTimeout = v.GetDuration(HTTPS_READ_TIMEOUT_KEY)
	hc.WriteTimeout = v.GetDuration(HTTPS_WRITE_TIMEOUT_KEY)
	hc.IdleTimeout = v.GetDuration(HTTPS_IDLE_TIMEOUT_KEY)
	hc.MaxIdleWorkerDuration = v.GetDuration(HTTPS_MAX_IDLE_WORKER_DURATION_KEY)
	hc.MaxReqBodySize = v.GetInt(HTTPS_MAX_REQ_BODY_SIZE_KEY)
	hc.ReadBufferSize = v.GetInt(HTTPS_READ_BUFFER_SIZE_KEY)
	hc.WriteBufferSize = v.GetInt(HTTPS_WRITE_BUFFER_SIZE_KEY)
	hc.Concurrency = v.GetInt(HTTPS_CONCURRENCY_KEY)
	hc.MaxConnsPerIP = v.GetInt(HTTPS_MAX_CONNS_PER_IP_KEY)
	hc.MaxRequestsPerConn = v.GetInt(HTTPS_MAX_REQUESTS_PER_CONN_KEY)
	hc.TCPKeepAlivePeriod = v.GetDuration(HTTPS_TCP_KEEPALIVE_PERIOD_KEY)
}

func (hc *HTTPServerConfig) SetDefaults(v *viper.Viper) {
	v.SetDefault(HTTPS_READ_TIMEOUT_KEY, time.Second*5)
	v.SetDefault(HTTPS_WRITE_TIMEOUT_KEY, time.Second*10)
	v.SetDefault(HTTPS_IDLE_TIMEOUT_KEY, time.Second*30)
	v.SetDefault(HTTPS_MAX_IDLE_WORKER_DURATION_KEY, time.Second*10)
	v.SetDefault(HTTPS_MAX_REQ_BODY_SIZE_KEY, 4*1024*1024)
	v.SetDefault(HTTPS_READ_BUFFER_SIZE_KEY, 4*1024)
	v.SetDefault(HTTPS_WRITE_BUFFER_SIZE_KEY, 4*1024)
	v.SetDefault(HTTPS_CONCURRENCY_KEY, 256*1024)
	v.SetDefault(HTTPS_MAX_CONNS_PER_IP_KEY, 100)
	v.SetDefault(HTTPS_MAX_REQUESTS_PER_CONN_KEY, 1000)
	v.SetDefault(HTTPS_TCP_KEEPALIVE_PERIOD_KEY, time.Minute*3)
}
