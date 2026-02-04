package application

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/dnonakolesax/noted-auth/internal/metrics"
)

type Metrics struct {
	TokenGetMetrics      *metrics.HTTPRequestMetrics
	SessionGetMetrics    *metrics.HTTPRequestMetrics
	SessionDeleteMetrics *metrics.HTTPRequestMetrics

	Reg *prometheus.Registry
}

func (a *App) SetupMetrics() {
	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	tokenRequestMetrics := metrics.NewHTTPRequestMetrics(reg, "keycloak_token_post")
	sessionGetMetrics := metrics.NewHTTPRequestMetrics(reg, "keycloak_session_get")
	sessionDeleteMetrics := metrics.NewHTTPRequestMetrics(reg, "keycloak_session_delete")

	a.metrics = &Metrics{
		TokenGetMetrics:      tokenRequestMetrics,
		SessionGetMetrics:    sessionGetMetrics,
		SessionDeleteMetrics: sessionDeleteMetrics,
		Reg:                  reg,
	}
}
