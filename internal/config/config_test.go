package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/vaultpipe/internal/config"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return p
}

func TestLoad_YAML(t *testing.T) {
	path := writeTemp(t, "cfg.yaml", `
vault:
  address: http://127.0.0.1:8200
  token: s.test
secrets:
  - path: secret/data/app
    prefix: APP_
inject:
  overwrite: true
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("address = %q, want http://127.0.0.1:8200", cfg.Vault.Address)
	}
	if len(cfg.Secrets) != 1 || cfg.Secrets[0].Prefix != "APP_" {
		t.Errorf("unexpected secrets: %+v", cfg.Secrets)
	}
	if !cfg.Inject.Overwrite {
		t.Error("expected overwrite = true")
	}
}

func TestLoad_JSON(t *testing.T) {
	path := writeTemp(t, "cfg.json", `{
  "vault": {"address": "http://vault:8200", "token": "tok"},
  "secrets": [{"path": "kv/data/svc", "prefix": ""}]
}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Token != "tok" {
		t.Errorf("token = %q, want tok", cfg.Vault.Token)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	path := writeTemp(t, "cfg.yaml", `
vault:
  address: http://original:8200
  token: original-token
`)
	t.Setenv("VAULT_ADDR", "http://override:8200")
	t.Setenv("VAULT_TOKEN", "override-token")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://override:8200" {
		t.Errorf("address not overridden: %q", cfg.Vault.Address)
	}
	if cfg.Vault.Token != "override-token" {
		t.Errorf("token not overridden: %q", cfg.Vault.Token)
	}
}

func TestLoad_MissingAddress(t *testing.T) {
	path := writeTemp(t, "cfg.yaml", `vault:\n  token: tok\n`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing address")
	}
}

func TestLoad_EmptyPath_EnvOnly(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://env:8200")
	t.Setenv("VAULT_TOKEN", "env-token")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://env:8200" {
		t.Errorf("unexpected address: %q", cfg.Vault.Address)
	}
}
