// Package cache implements a thread-safe, TTL-based in-memory cache for
// Vault secret responses.
//
// vaultpipe uses this cache to avoid redundant round-trips to the Vault
// server when the same secret path is requested multiple times within a
// short window. Each entry is stored with a configurable expiration time
// (default: 5 minutes) and is evicted lazily on the next read.
//
// Usage:
//
//	c := cache.New(cache.WithTTL(2 * time.Minute))
//	c.Set("secret/data/myapp", secrets)
//	if vals, ok := c.Get("secret/data/myapp"); ok {
//	    // use vals
//	}
//
// The cache does not run a background goroutine for eviction; stale entries
// are detected and ignored at read time.
package cache
