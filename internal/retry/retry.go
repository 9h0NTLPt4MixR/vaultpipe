// Package retry provides a simple exponential back-off retry helper used
// when communicating with HashiCorp Vault.  It is intentionally small and
// carries no external dependencies beyond the standard library.
package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// ErrMaxAttempts is returned when all retry attempts have been exhausted.
var ErrMaxAttempts = errors.New("retry: max attempts reached")

// Config holds the parameters that control retry behaviour.
type Config struct {
	// MaxAttempts is the total number of tries (including the first call).
	MaxAttempts int
	// BaseDelay is the initial back-off duration before the first retry.
	BaseDelay time.Duration
	// MaxDelay caps the exponential growth of the back-off.
	MaxDelay time.Duration
	// Factor is the multiplier applied to the delay on each attempt (≥1).
	Factor float64
}

// DefaultConfig returns a Config with sensible defaults for Vault calls.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 5,
		BaseDelay:   200 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Factor:      2.0,
	}
}

// Do calls fn repeatedly until it returns nil, the context is cancelled, or
// MaxAttempts is exhausted.  It returns ErrMaxAttempts (wrapping the last
// error) when all tries fail.
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}
	if cfg.Factor < 1 {
		cfg.Factor = 1
	}

	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		delay := time.Duration(float64(cfg.BaseDelay) * math.Pow(cfg.Factor, float64(attempt)))
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return errors.Join(ErrMaxAttempts, lastErr)
}
