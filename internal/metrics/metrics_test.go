package metrics_test

import (
	"testing"

	"github.com/your-org/vaultpipe/internal/metrics"
)

func TestNew_ZeroValues(t *testing.T) {
	m := metrics.New()
	snap := m.Snapshot()

	for key, val := range snap {
		if val != 0 {
			t.Errorf("expected counter %q to be 0, got %d", key, val)
		}
	}
}

func TestCounter_Inc(t *testing.T) {
	m := metrics.New()

	m.SecretFetches.Inc()
	m.SecretFetches.Inc()
	m.FetchErrors.Inc()

	if got := m.SecretFetches.Value(); got != 2 {
		t.Errorf("SecretFetches: want 2, got %d", got)
	}
	if got := m.FetchErrors.Value(); got != 1 {
		t.Errorf("FetchErrors: want 1, got %d", got)
	}
}

func TestSnapshot_AllKeys(t *testing.T) {
	m := metrics.New()
	expectedKeys := []string{
		"secret_fetches", "fetch_errors",
		"cache_hits", "cache_misses",
		"injections_ok", "injection_errors",
		"renewals", "renewal_errors",
	}

	snap := m.Snapshot()
	for _, k := range expectedKeys {
		if _, ok := snap[k]; !ok {
			t.Errorf("Snapshot missing key %q", k)
		}
	}
}

func TestSnapshot_ReflectsIncrements(t *testing.T) {
	m := metrics.New()
	m.CacheHits.Inc()
	m.CacheHits.Inc()
	m.CacheHits.Inc()
	m.CacheMisses.Inc()

	snap := m.Snapshot()
	if snap["cache_hits"] != 3 {
		t.Errorf("cache_hits: want 3, got %d", snap["cache_hits"])
	}
	if snap["cache_misses"] != 1 {
		t.Errorf("cache_misses: want 1, got %d", snap["cache_misses"])
	}
}

func TestSnapshot_IsolatedInstances(t *testing.T) {
	a := metrics.New()
	b := metrics.New()

	a.Renewals.Inc()

	if b.Renewals.Value() != 0 {
		t.Error("instances should not share state")
	}
}
