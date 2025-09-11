package configs

import (
	"time"

	"github.com/spf13/viper"
)

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

func NewHTTPServerConfig(v *viper.Viper) HTTPServerConfig {
	return HTTPServerConfig{
		ReadTimeout:           v.GetDuration("http-server.read-timeout"),
		WriteTimeout:          v.GetDuration("http-server.write-timeout"),
		IdleTimeout:           v.GetDuration("http-server.idle-timeout"),
		MaxIdleWorkerDuration: v.GetDuration("http-server.max-idle-worker-duration"),
		MaxReqBodySize:        v.GetInt("http-server.max-req-body-size"),
		ReadBufferSize:        v.GetInt("http-server.read-buffer-size"),
		WriteBufferSize:       v.GetInt("http-server.write-buffer-size"),
		Concurrency:           v.GetInt("http-server.concurrency"),
		MaxConnsPerIP:         v.GetInt("http-server.max-conns-per-ip"),
		MaxRequestsPerConn:    v.GetInt("http-server.max-requests-per-conn"),
		TCPKeepAlivePeriod:    v.GetDuration("http-server.tcp-keepalive-period"),
	}
}
