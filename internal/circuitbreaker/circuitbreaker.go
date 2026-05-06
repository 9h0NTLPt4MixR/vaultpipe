// Package circuitbreaker provides a simple circuit breaker implementation
// for protecting calls to external systems such as HashiCorp Vault.
//
// The circuit breaker transitions between three states:
//
//   - Closed: requests flow normally; failures are counted.
//   - Open: requests are rejected immediately without calling the target.
//   - Half-Open: a single probe request is allowed through to test recovery.
//
// Use New to create a breaker with sensible defaults, and optionally tune it
// with WithThreshold, WithTimeout, and WithOnStateChange.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker.
type State int

const (
	// StateClosed allows requests to pass through normally.
	StateClosed State = iota
	// StateOpen rejects all requests without invoking the target.
	StateOpen
	// StateHalfOpen allows a single probe request to test recovery.
	StateHalfOpen
)

// String returns a human-readable label for the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrOpen is returned when a call is rejected because the circuit is open.
var ErrOpen = errors.New("circuit breaker is open")

// OnStateChangeFunc is called whenever the breaker transitions between states.
type OnStateChangeFunc func(from, to State)

// Option configures a CircuitBreaker.
type Option func(*CircuitBreaker)

// WithThreshold sets the number of consecutive failures required to open the
// circuit. Defaults to 5.
func WithThreshold(n int) Option {
	return func(cb *CircuitBreaker) {
		if n > 0 {
			cb.threshold = n
		}
	}
}

// WithTimeout sets how long the circuit stays open before moving to half-open.
// Defaults to 30 seconds.
func WithTimeout(d time.Duration) Option {
	return func(cb *CircuitBreaker) {
		if d > 0 {
			cb.timeout = d
		}
	}
}

// WithOnStateChange registers a callback invoked on every state transition.
func WithOnStateChange(fn OnStateChangeFunc) Option {
	return func(cb *CircuitBreaker) {
		cb.onStateChange = fn
	}
}

// CircuitBreaker guards calls to an external dependency.
type CircuitBreaker struct {
	mu            sync.Mutex
	state         State
	failures      int
	threshold     int
	timeout       time.Duration
	openedAt      time.Time
	onStateChange OnStateChangeFunc
}

// New returns a CircuitBreaker with default settings, optionally modified by
// the provided options.
func New(opts ...Option) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:     StateClosed,
		threshold: 5,
		timeout:   30 * time.Second,
	}
	for _, o := range opts {
		o(cb)
	}
	return cb
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentState()
}

// currentState resolves the effective state, transitioning from open to
// half-open when the timeout has elapsed. Must be called with mu held.
func (cb *CircuitBreaker) currentState() State {
	if cb.state == StateOpen && time.Since(cb.openedAt) >= cb.timeout {
		cb.transition(StateHalfOpen)
	}
	return cb.state
}

// Do executes fn if the circuit is closed or half-open. A successful call
// resets the breaker; a failed call increments the failure counter and may
// open the circuit.
func (cb *CircuitBreaker) Do(fn func() error) error {
	cb.mu.Lock()
	state := cb.currentState()
	if state == StateOpen {
		cb.mu.Unlock()
		return ErrOpen
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
		return err
	}
	cb.reset()
	return nil
}

// recordFailure increments the failure counter and opens the circuit if the
// threshold is reached. Must be called with mu held.
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	if cb.state == StateHalfOpen || cb.failures >= cb.threshold {
		cb.openedAt = time.Now()
		cb.transition(StateOpen)
	}
}

// reset clears the failure counter and closes the circuit.
// Must be called with mu held.
func (cb *CircuitBreaker) reset() {
	cb.failures = 0
	cb.transition(StateClosed)
}

// transition moves the breaker to the given state and fires the callback if
// the state actually changed. Must be called with mu held.
func (cb *CircuitBreaker) transition(next State) {
	if cb.state == next {
		return
	}
	prev := cb.state
	cb.state = next
	if cb.onStateChange != nil {
		cb.onStateChange(prev, next)
	}
}
