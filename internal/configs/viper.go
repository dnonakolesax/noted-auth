package configs

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func Load(path string, v *viper.Viper) error {
	setDefaults(v)
	v.AddConfigPath(path)

	v.SetConfigFile("./" + path + "/config.yaml")
	v.SetConfigType("yaml")
	err := v.MergeInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Error("Config file not found: %s", path)
			return nil
		}
		return fmt.Errorf("failed to merge config: %w", err)
	}

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	err = v.MergeInConfig()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Error("Config file not found: %s", path)
			return nil
		}
		return fmt.Errorf("failed to merge config: %w", err)
	}

	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("service.allowed-redirect", "https://dnk33.com/")
	v.SetDefault("service.port", 8080)
	v.SetDefault("service.grpc-port", 8081)
	v.SetDefault("service.metrics-port", 8082)
	v.SetDefault("service.base-path", "/openid-connect")
	v.SetDefault("service.auth-timeout", time.Duration(5)*time.Minute)
	//v.SetDefault("service.shutdown_timeout", "30s")
	//v.SetDefault("service.env", "development")

	v.SetDefault("postgres.address", "postgres://")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.database", "keycloak")
	v.SetDefault("postgres.user", "keycloak-selector")
	// v.SetDefault("postgres.timeout", "10s")
	// v.SetDefault("postgres.max_connections", 100)

	v.SetDefault("realm.base-url", "http://localhost:8080/realms/noted/protocol/openid-connect/")
	v.SetDefault("realm.client-id", "admin-cli")
	v.SetDefault("realm.client-secret", "zizipabeda")
	v.SetDefault("realm.redirect-url", "http://localhost:8080/callback")
	v.SetDefault("realm.te-timeout", time.Duration(10)*time.Second)
	v.SetDefault("realm.state-length", 32)
	v.SetDefault("realm.id", "00000000-0000-0000-0000-000000000000")

	// Redis defaults
	v.SetDefault("redis.address", "redis://")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "")
	// v.SetDefault("redis.db", 0)
	// v.SetDefault("redis.timeout", "5s")
}
