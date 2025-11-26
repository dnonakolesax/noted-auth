package application

import "sync/atomic"

type HealthChecks struct {
	Postgres *atomic.Bool
	Redis    *atomic.Bool
	Keycloak *atomic.Bool
	Vault    *atomic.Bool
}

func (a *App) InitHealthchecks() {
	a.health = &HealthChecks{
		Postgres: &atomic.Bool{},
		Redis:    &atomic.Bool{},
		Keycloak: &atomic.Bool{},
		Vault:    &atomic.Bool{},
	}
}
