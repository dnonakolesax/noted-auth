package configs

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dnonakolesax/viper"
)

type configurable interface {
	SetDefaults(v *viper.Viper)
	Load(v *viper.Viper)
}

func Load(path string, v *viper.Viper, configs ...configurable) error {
	for _, cfg := range configs {
		cfg.SetDefaults(v)
	}

	v.AddConfigPath(path)
	v.AddConfigPath("./")

	v.SetConfigName("config")
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

	for _, cfg := range configs {
		cfg.Load(v)
	}

	return nil
}
