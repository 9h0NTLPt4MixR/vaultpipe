// Package metrics provides lightweight runtime counters for vaultpipe
// operations such as secret fetches, cache hits, and injection events.
package metrics

import (
	"sync"
	"sync/atomic"
)

// Counter is a monotonically increasing uint64 counter.
type Counter struct{ v uint64 }

// Inc increments the counter by 1.
func (c *Counter) Inc() { atomic.AddUint64(&c.v, 1) }

// Value returns the current counter value.
func (c *Counter) Value() uint64 { return atomic.LoadUint64(&c.v) }

// Metrics holds all runtime counters for vaultpipe.
type Metrics struct {
	mu sync.RWMutex

	// Vault operations
	SecretFetches  Counter
	FetchErrors    Counter

	// Cache operations
	CacheHits   Counter
	CacheMisses Counter

	// Injection operations
	InjectionsOK    Counter
	InjectionErrors Counter

	// Renewal operations
	Renewals      Counter
	RenewalErrors Counter
}

// New returns a new zero-value Metrics instance.
func New() *Metrics {
	return &Metrics{}
}

// Snapshot returns a point-in-time copy of all counter values as a map.
func (m *Metrics) Snapshot() map[string]uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]uint64{
		"secret_fetches":   m.SecretFetches.Value(),
		"fetch_errors":     m.FetchErrors.Value(),
		"cache_hits":       m.CacheHits.Value(),
		"cache_misses":     m.CacheMisses.Value(),
		"injections_ok":    m.InjectionsOK.Value(),
		"injection_errors": m.InjectionErrors.Value(),
		"renewals":         m.Renewals.Value(),
		"renewal_errors":   m.RenewalErrors.Value(),
	}
}
