package configs

import (
	"strings"
	"log/slog"
	"fmt"
	"github.com/spf13/viper"
)

func Load(path string, v *viper.Viper) error {
	setDefaults(v)
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.AddConfigPath(path)
	v.SetConfigName("config")
	err := v.MergeInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Error("Config file not found: %s", path)
			return nil
		}
		return fmt.Errorf("failed to merge config: %w", err)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
		slog.Warn("No config file found, using defaults and environment variables")
	}

	return nil
}

func setDefaults(v *viper.Viper) { 
	v.SetDefault("app.allowed-redirect", "https://dnk33.com/")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.base-path", "/openid-connect")
	v.SetDefault("app.auth-timeout", 5)
	//v.SetDefault("app.shutdown_timeout", "30s")
	//v.SetDefault("app.env", "development")

	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.database", "keycloak")
	v.SetDefault("postgres.user", "keycloak-selector")
	// v.SetDefault("postgres.timeout", "10s")
	// v.SetDefault("postgres.max_connections", 100)

	v.SetDefault("realm.base-url", "http://localhost:8080/realms/noted/protocol/openid-connect/")
	v.SetDefault("realm.client-id", "admin-cli")
	v.SetDefault("realm.client-secret", "zizipabeda")
	v.SetDefault("realm.redirect-url", "http://localhost:8080/callback")
	v.SetDefault("realm.te-timeout", 10)
	v.SetDefault("realm.state-length", 32)
	

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	// v.SetDefault("redis.db", 0)
	// v.SetDefault("redis.timeout", "5s")
}