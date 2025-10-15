package model

type HealthcheckDTO struct { //nolint:recvcheck // autogen issues
	RedisAlive    bool `json:"redis_alive"`
	PostgresAlive bool `json:"postgres_alive"`
	KeycloakAlive bool `json:"keycloak_alive"`
	VaultAlive    bool `json:"vault_alive"`
}
