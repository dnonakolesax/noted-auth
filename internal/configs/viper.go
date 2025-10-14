package configs

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dnonakolesax/viper"

	"github.com/dnonakolesax/noted-auth/internal/consts"
)

type configurable interface {
	SetDefaults(v *viper.Viper)
	Load(v *viper.Viper)
}

func Load(path string, v *viper.Viper, logger *slog.Logger, configs ...configurable) error {
	for _, cfg := range configs {
		cfg.SetDefaults(v)
	}

	v.AddConfigPath(path)
	v.AddConfigPath("./")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	err := v.MergeInConfig()

	if err != nil {
		var vErr viper.ConfigFileNotFoundError
		if errors.As(err, &vErr) {
			logger.Error("Config file not found yaml")
			return nil
		}
		logger.Error("Failed to merge yaml config", slog.String(consts.ErrorLoggerKey, err.Error()))
		return fmt.Errorf("failed to merge config: %w", err)
	}

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	err = v.MergeInConfig()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// creds := &viper.RemoteCredentials{
	// 	AuthType: "userpass",
	// 	Login: "dunkelheit",
	// 	Password: "dunkelheit",
	// }
	// err = v.AddRemoteProvider("vault", "http://192.168.80.3:8200", "sample/zizipabeda:sample", creds)

	// if err != nil {
	// 	return fmt.Errorf("Error adding remote provider %s", err)
	// }

	// err = v.ReadRemoteConfig()

	// if err != nil {
	// 	return fmt.Errorf("Error reading remote config: %s", err)
	// }

	if err != nil {
		var vErr viper.ConfigFileNotFoundError
		if errors.As(err, &vErr) {
			logger.Error("Config file not found env")
			return nil
		}
		logger.Error("Failed to merge dotenv config", slog.String(consts.ErrorLoggerKey, err.Error()))
		return fmt.Errorf("failed to merge config: %w", err)
	}

	for _, cfg := range configs {
		cfg.Load(v)
	}

	return nil
}
