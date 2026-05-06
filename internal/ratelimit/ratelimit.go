// Package ratelimit provides a token-bucket rate limiter for controlling
// the frequency of secret fetch requests made to HashiCorp Vault.
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Limiter controls the rate at which operations are permitted.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	clock    func() time.Time
}

// Option configures a Limiter.
type Option func(*Limiter)

// WithRate sets the number of tokens replenished per second.
// Defaults to 1.0.
func WithRate(r float64) Option {
	return func(l *Limiter) {
		l.rate = r
	}
}

// WithBurst sets the maximum token capacity (burst size).
// Defaults to 1.
func WithBurst(b float64) Option {
	return func(l *Limiter) {
		l.max = b
		l.tokens = b
	}
}

// New creates a Limiter with the given options.
func New(opts ...Option) (*Limiter, error) {
	l := &Limiter{
		rate:     1.0,
		max:      1.0,
		tokens:   1.0,
		lastTick: time.Now(),
		clock:    time.Now,
	}
	for _, o := range opts {
		o(l)
	}
	if l.rate <= 0 {
		return nil, fmt.Errorf("ratelimit: rate must be positive, got %v", l.rate)
	}
	if l.max <= 0 {
		return nil, fmt.Errorf("ratelimit: burst must be positive, got %v", l.max)
	}
	return l, nil
}

// Wait blocks until a token is available or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := l.clock()
		elapsed := now.Sub(l.lastTick).Seconds()
		l.tokens += elapsed * l.rate
		if l.tokens > l.max {
			l.tokens = l.max
		}
		l.lastTick = now

		if l.tokens >= 1.0 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		// Calculate wait duration for next token.
		wait := time.Duration((1.0-l.tokens)/l.rate*1000) * time.Millisecond
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired.
func (l *Limiter) TryAcquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clock()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}
	l.lastTick = now
	if l.tokens >= 1.0 {
		l.tokens--
		return true
	}
	return false
}
