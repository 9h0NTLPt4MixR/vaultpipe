package vault

import (
	"context"
	"fmt"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

// Config holds configuration for the Vault client.
type Config struct {
	Address   string
	Token     string
	Namespace string
}

// Client wraps the HashiCorp Vault API client.
type Client struct {
	api *vaultapi.Client
}

// NewClient creates a new Vault client from the provided config.
// If config fields are empty, it falls back to environment variables
// (VAULT_ADDR, VAULT_TOKEN, VAULT_NAMESPACE).
func NewClient(cfg Config) (*Client, error) {
	addr := cfg.Address
	if addr == "" {
		addr = os.Getenv("VAULT_ADDR")
	}
	if addr == "" {
		addr = "http://127.0.0.1:8200"
	}

	token := cfg.Token
	if token == "" {
		token = os.Getenv("VAULT_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("vault token is required: set VAULT_TOKEN or provide Config.Token")
	}

	apicfg := vaultapi.DefaultConfig()
	apicfg.Address = addr

	apiClient, err := vaultapi.NewClient(apicfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	apiClient.SetToken(token)

	ns := cfg.Namespace
	if ns == "" {
		ns = os.Getenv("VAULT_NAMESPACE")
	}
	if ns != "" {
		apiClient.SetNamespace(ns)
	}

	return &Client{api: apiClient}, nil
}

// GetSecrets reads a KV v2 secret at the given path and returns
// the key/value pairs as a map[string]string.
func (c *Client) GetSecrets(ctx context.Context, path string) (map[string]string, error) {
	secret, err := c.api.KVv2("").Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at path %q: %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path %q", path)
	}

	result := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("secret key %q at path %q is not a string value", k, path)
		}
		result[k] = str
	}
	return result, nil
}
