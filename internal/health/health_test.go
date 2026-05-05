package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/youorg/vaultpipe/internal/health"
)

func TestNew_DefaultsToNotReady(t *testing.T) {
	c := health.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	c.Handler()(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var s health.Status
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.OK {
		t.Error("expected ok=false")
	}
	if s.VaultReady {
		t.Error("expected vault_ready=false")
	}
	if s.Message == "" {
		t.Error("expected non-empty message when unhealthy")
	}
}

func TestSetVaultReady_Returns200(t *testing.T) {
	c := health.New()
	c.SetVaultReady(true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	c.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var s health.Status
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !s.OK {
		t.Error("expected ok=true")
	}
	if s.Message != "" {
		t.Errorf("expected empty message, got %q", s.Message)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	c := health.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)

	c.Handler()(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestSetVaultReady_Transition(t *testing.T) {
	c := health.New()
	c.SetVaultReady(true)
	c.SetVaultReady(false)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	c.Handler()(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 after transition to not-ready, got %d", rec.Code)
	}
}
