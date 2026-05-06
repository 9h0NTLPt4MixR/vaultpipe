// Package backoff provides configurable exponential back-off strategies
// used when retrying transient failures against Vault or downstream services.
package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Strategy computes the delay before the next retry attempt.
type Strategy interface {
	// Next returns the duration to wait before attempt n (0-indexed).
	Next(attempt int) time.Duration
}

// Config holds parameters for an exponential back-off with optional jitter.
type Config struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	Jitter     bool
}

// DefaultConfig returns sensible defaults suitable for Vault API calls.
func DefaultConfig() Config {
	return Config{
		BaseDelay:  200 * time.Millisecond,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

type exponential struct {
	cfg Config
}

// New returns a Strategy built from cfg. If Multiplier is <= 1 it is set to 2.
func New(cfg Config) Strategy {
	if cfg.Multiplier <= 1 {
		cfg.Multiplier = 2
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 200 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 30 * time.Second
	}
	return &exponential{cfg: cfg}
}

// Next computes delay = min(base * multiplier^attempt, max).
// When Jitter is enabled a random value in [0, delay] is added and the result
// is capped at MaxDelay again.
func (e *exponential) Next(attempt int) time.Duration {
	delay := float64(e.cfg.BaseDelay) * math.Pow(e.cfg.Multiplier, float64(attempt))
	if delay > float64(e.cfg.MaxDelay) {
		delay = float64(e.cfg.MaxDelay)
	}
	if e.cfg.Jitter {
		// nolint:gosec — not used for cryptographic purposes
		delay += rand.Float64() * delay
		if delay > float64(e.cfg.MaxDelay) {
			delay = float64(e.cfg.MaxDelay)
		}
	}
	return time.Duration(delay)
}
