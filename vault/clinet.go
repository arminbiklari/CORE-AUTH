package vault

import (
	"context"
	"crypto/tls"
	"net/http"

	"core-auth/config"

	vault "github.com/hashicorp/vault/api"
)

type VaultClient struct {
	client *vault.Client
	config *config.Config
}

func NewVaultClient(cfg *config.Config) (*VaultClient, error) {
	vaultConfig := vault.DefaultConfig()
	vaultConfig.Address = cfg.Vault.Addr

	if cfg.Vault.SkipVerify {
		vaultConfig.HttpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}

	client, err := vault.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}

	// Set token for authentication
	client.SetToken(cfg.Vault.Token)

	return &VaultClient{
		client: client,
		config: cfg,
	}, nil
}

// GetDatabaseCredentials retrieves dynamic database credentials from Vault
func (v *VaultClient) GetDatabaseCredentials(ctx context.Context) (string, string, error) {
	secret, err := v.client.Logical().ReadWithContext(ctx, v.config.Vault.RolePath)
	if err != nil {
		return "", "", err
	}

	username, ok := secret.Data["username"].(string)
	if !ok {
		return "", "", err
	}

	password, ok := secret.Data["password"].(string)
	if !ok {
		return "", "", err
	}

	return username, password, nil
}