package vault_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultpipe/internal/vault"
)

func newMockVaultServer(t *testing.T, path string, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/v1/secret/data/" + path
		if !strings.HasSuffix(r.URL.Path, expected) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"data":     data,
				"metadata": map[string]interface{}{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestNewClient_MissingToken(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "")
	_, err := vault.NewClient(vault.Config{})
	if err == nil {
		t.Fatal("expected error when token is missing, got nil")
	}
}

func TestNewClient_WithToken(t *testing.T) {
	srv := newMockVaultServer(t, "myapp/prod", map[string]interface{}{
		"DB_PASSWORD": "s3cr3t",
	})
	defer srv.Close()

	client, err := vault.NewClient(vault.Config{
		Address: srv.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetSecrets_Success(t *testing.T) {
	expected := map[string]interface{}{
		"API_KEY":     "abc123",
		"DB_PASSWORD": "hunter2",
	}
	srv := newMockVaultServer(t, "myapp/prod", expected)
	defer srv.Close()

	client, err := vault.NewClient(vault.Config{
		Address: srv.URL,
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	secrets, err := client.GetSecrets(context.Background(), "myapp/prod")
	if err != nil {
		t.Fatalf("GetSecrets returned error: %v", err)
	}

	for k, v := range expected {
		got, ok := secrets[k]
		if !ok {
			t.Errorf("missing key %q in result", k)
			continue
		}
		if got != v.(string) {
			t.Errorf("key %q: got %q, want %q", k, got, v)
		}
	}
}
