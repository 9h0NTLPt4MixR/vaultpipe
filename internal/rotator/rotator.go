// Package rotator provides secret rotation coordination for vaultpipe.
// It detects when a Vault lease or token is nearing expiry and triggers
// a re-fetch cycle so downstream processes always receive fresh secrets.
package rotator

import (
	"context"
	"sync"
	"time"

	"github.com/your-org/vaultpipe/internal/logger"
)

// RotateFunc is called when secrets should be rotated.
type RotateFunc func(ctx context.Context) error

// Option configures a Rotator.
type Option func(*Rotator)

// Rotator periodically checks whether secrets need rotation and invokes
// a caller-supplied RotateFunc when the renewal window is reached.
type Rotator struct {
	interval  time.Duration
	threshold time.Duration
	log       *logger.Logger
	mu        sync.Mutex
	lastRotation time.Time
}

// WithInterval sets how often the rotator checks for expiry.
func WithInterval(d time.Duration) Option {
	return func(r *Rotator) {
		if d > 0 {
			r.interval = d
		}
	}
}

// WithThreshold sets how far before expiry a rotation should be triggered.
func WithThreshold(d time.Duration) Option {
	return func(r *Rotator) {
		if d > 0 {
			r.threshold = d
		}
	}
}

// New creates a Rotator with sensible defaults.
func New(log *logger.Logger, opts ...Option) *Rotator {
	r := &Rotator{
		interval:  30 * time.Second,
		threshold: 5 * time.Minute,
		log:       log,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Run starts the rotation loop. It blocks until ctx is cancelled.
func (r *Rotator) Run(ctx context.Context, expiry time.Time, fn RotateFunc) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:
			if now.After(expiry.Add(-r.threshold)) {
				r.log.Info("rotation threshold reached, rotating secrets")
				if err := fn(ctx); err != nil {
					r.log.Error("rotation failed", "error", err)
					continue
				}
				r.mu.Lock()
				r.lastRotation = now
				r.mu.Unlock()
			}
		}
	}
}

// LastRotation returns the time of the most recent successful rotation.
func (r *Rotator) LastRotation() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastRotation
}
