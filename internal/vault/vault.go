package vault

import (
	"context"
	"log/slog"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/consts"
)

func SetupVault(cfg *configs.VaultConfig, logger *slog.Logger) *vault.Client {
	vClient, err := vault.New(vault.WithAddress(cfg.Address))
	if err != nil {
		logger.Error("Coulndt create vault client", slog.String(consts.ErrorLoggerKey, err.Error()))
		panic(err)
	}
	resp, err := vClient.Auth.UserpassLogin(context.Background(), cfg.Login, schema.UserpassLoginRequest{
		Password: cfg.Password,
	})
	if err != nil {
		logger.Error("Coulndt create vault client", slog.String(consts.ErrorLoggerKey, err.Error()))
		panic(err)
	}
	err = vClient.SetToken(resp.Auth.ClientToken)
	if err != nil {
		logger.Error("Coulndt set vault token", slog.String(consts.ErrorLoggerKey, err.Error()))
		panic(err)
	}
	return vClient
}
