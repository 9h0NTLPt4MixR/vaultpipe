// Package config provides configuration loading for vaultpipe.
// It supports loading from environment variables and YAML/JSON config files.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level vaultpipe configuration.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"   json:"vault"`
	Secrets []SecretMount  `yaml:"secrets"  json:"secrets"`
	Inject  InjectConfig  `yaml:"inject"  json:"inject"`
}

// VaultConfig contains connection settings for HashiCorp Vault.
type VaultConfig struct {
	Address   string `yaml:"address"   json:"address"`
	Token     string `yaml:"token"     json:"token"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// SecretMount maps a Vault secret path to an optional env prefix.
type SecretMount struct {
	Path   string `yaml:"path"   json:"path"`
	Prefix string `yaml:"prefix" json:"prefix"`
}

// InjectConfig controls injection behaviour.
type InjectConfig struct {
	Overwrite bool `yaml:"overwrite" json:"overwrite"`
}

// Load reads configuration from the given file path.
// Supported formats: .yaml / .yml and .json.
// Environment variables override file values:
//   VAULT_ADDR, VAULT_TOKEN, VAULT_NAMESPACE
func Load(path string) (*Config, error) {
	cfg := &Config{}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("config: reading file %q: %w", path, err)
		}

		switch {
		case strings.HasSuffix(path, ".json"):
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("config: parsing JSON: %w", err)
			}
		default: // yaml / yml
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("config: parsing YAML: %w", err)
			}
		}
	}

	// Environment variable overrides.
	if v := os.Getenv("VAULT_ADDR"); v != "" {
		cfg.Vault.Address = v
	}
	if v := os.Getenv("VAULT_TOKEN"); v != "" {
		cfg.Vault.Token = v
	}
	if v := os.Getenv("VAULT_NAMESPACE"); v != "" {
		cfg.Vault.Namespace = v
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("config: vault.address is required")
	}
	if c.Vault.Token == "" {
		return fmt.Errorf("config: vault.token is required (or set VAULT_TOKEN)")
	}
	return nil
}
