// Package cache provides a simple TTL-based in-memory cache for Vault secrets
// to reduce the number of requests made to the Vault server.
package cache

import (
	"sync"
	"time"
)

// Entry holds a cached secret value along with its expiration time.
type Entry struct {
	Secrets   map[string]string
	ExpiresAt time.Time
}

// Cache is a thread-safe TTL cache for Vault secret paths.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]Entry
	ttl     time.Duration
	nowFunc func() time.Time
}

// Option is a functional option for configuring a Cache.
type Option func(*Cache)

// WithTTL sets the time-to-live for cache entries.
func WithTTL(ttl time.Duration) Option {
	return func(c *Cache) {
		c.ttl = ttl
	}
}

// New creates a new Cache with the given options.
// The default TTL is 5 minutes.
func New(opts ...Option) *Cache {
	c := &Cache{
		items:   make(map[string]Entry),
		ttl:     5 * time.Minute,
		nowFunc: time.Now,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Set stores secrets for the given path in the cache.
func (c *Cache) Set(path string, secrets map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	copy := make(map[string]string, len(secrets))
	for k, v := range secrets {
		copy[k] = v
	}
	c.items[path] = Entry{
		Secrets:   copy,
		ExpiresAt: c.nowFunc().Add(c.ttl),
	}
}

// Get retrieves secrets for the given path. Returns the secrets and true if
// found and not expired, otherwise returns nil and false.
func (c *Cache) Get(path string) (map[string]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[path]
	if !ok || c.nowFunc().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Secrets, true
}

// Invalidate removes the cached entry for the given path.
func (c *Cache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, path)
}

// Flush removes all entries from the cache.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Entry)
}
