package configs

import (
	"os"
)

type VaultConfig struct {
	Address  string
	Login    string
	Password string
}

func NewVaultConfig() *VaultConfig {
	address := os.Getenv("VAULT_ADDRESS")
	login := os.Getenv("VAULT_LOGIN")
	password := os.Getenv("VAULT_PASSWORD")

	os.Setenv("VAULT_PASSWORD", "there is no spoon, dear hacker")

	return &VaultConfig{
		Address:  address,
		Login:    login,
		Password: password,
	}
}
