// Package watcher monitors Vault secret paths for changes and triggers
// re-injection into the running process environment when secrets rotate.
package watcher

import (
	"context"
	"sync"
	"time"

	"github.com/yourorg/vaultpipe/internal/logger"
)

// SecretFetcher retrieves secrets from a given path.
type SecretFetcher interface {
	GetSecrets(ctx context.Context, path string) (map[string]string, error)
}

// ChangeHandler is called when a secret path yields new values.
type ChangeHandler func(path string, secrets map[string]string)

// Watcher polls one or more Vault secret paths at a fixed interval and
// invokes a ChangeHandler when the returned values differ from the last
// known state.
type Watcher struct {
	fetcher  SecretFetcher
	log      *logger.Logger
	interval time.Duration
	paths    []string
	handler  ChangeHandler
	mu       sync.Mutex
	cache    map[string]map[string]string
}

// Option configures a Watcher.
type Option func(*Watcher)

// WithInterval sets the polling interval (default: 30s).
func WithInterval(d time.Duration) Option {
	return func(w *Watcher) {
		if d > 0 {
			w.interval = d
		}
	}
}

// WithPaths registers secret paths to watch.
func WithPaths(paths ...string) Option {
	return func(w *Watcher) {
		w.paths = append(w.paths, paths...)
	}
}

// New creates a Watcher with the provided fetcher, handler, and options.
func New(fetcher SecretFetcher, handler ChangeHandler, log *logger.Logger, opts ...Option) *Watcher {
	w := &Watcher{
		fetcher:  fetcher,
		log:      log,
		interval: 30 * time.Second,
		handler:  handler,
		cache:    make(map[string]map[string]string),
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.poll(ctx) // immediate first poll

	for {
		select {
		case <-ticker.C:
			w.poll(ctx)
		case <-ctx.Done():
			w.log.Info("watcher stopped")
			return
		}
	}
}

func (w *Watcher) poll(ctx context.Context) {
	for _, path := range w.paths {
		secrets, err := w.fetcher.GetSecrets(ctx, path)
		if err != nil {
			w.log.WithError(err).Warn("watcher: failed to fetch secrets", "path", path)
			continue
		}
		if w.changed(path, secrets) {
			w.log.Info("watcher: secrets changed", "path", path)
			w.store(path, secrets)
			w.handler(path, secrets)
		}
	}
}

func (w *Watcher) changed(path string, incoming map[string]string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	prev, ok := w.cache[path]
	if !ok {
		return true
	}
	if len(prev) != len(incoming) {
		return true
	}
	for k, v := range incoming {
		if prev[k] != v {
			return true
		}
	}
	return false
}

func (w *Watcher) store(path string, secrets map[string]string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	copy := make(map[string]string, len(secrets))
	for k, v := range secrets {
		copy[k] = v
	}
	w.cache[path] = copy
}
