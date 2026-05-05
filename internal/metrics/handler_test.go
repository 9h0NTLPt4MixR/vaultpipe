package metrics_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/vaultpipe/internal/metrics"
)

func TestHandler_GET_ReturnsJSON(t *testing.T) {
	m := metrics.New()
	m.SecretFetches.Inc()
	m.CacheHits.Inc()
	m.CacheHits.Inc()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	metrics.Handler(m)(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: want application/json, got %q", ct)
	}

	var snap map[string]uint64
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if snap["secret_fetches"] != 1 {
		t.Errorf("secret_fetches: want 1, got %d", snap["secret_fetches"])
	}
	if snap["cache_hits"] != 2 {
		t.Errorf("cache_hits: want 2, got %d", snap["cache_hits"])
	}
}

func TestHandler_POST_MethodNotAllowed(t *testing.T) {
	m := metrics.New()

	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rec := httptest.NewRecorder()

	metrics.Handler(m)(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("want 405, got %d", rec.Code)
	}
}

func TestHandler_ZeroMetrics(t *testing.T) {
	m := metrics.New()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	metrics.Handler(m)(rec, req)

	var snap map[string]uint64
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	for k, v := range snap {
		if v != 0 {
			t.Errorf("expected zero for %q, got %d", k, v)
		}
	}
}
