// Package audit provides structured audit logging for secret access events.
package audit

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// EventType classifies the kind of audit event.
type EventType string

const (
	EventSecretFetch    EventType = "secret.fetch"
	EventSecretInject   EventType = "secret.inject"
	EventSecretRenew    EventType = "secret.renew"
	EventSecretCacheHit EventType = "secret.cache_hit"
)

// Event represents a single auditable action.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Type      EventType `json:"type"`
	Path      string    `json:"path,omitempty"`
	Keys      []string  `json:"keys,omitempty"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// Logger writes audit events as newline-delimited JSON.
type Logger struct {
	w       io.Writer
	enabled bool
}

// Option configures an audit Logger.
type Option func(*Logger)

// WithWriter overrides the default output writer.
func WithWriter(w io.Writer) Option {
	return func(l *Logger) { l.w = w }
}

// WithEnabled toggles audit logging on or off.
func WithEnabled(enabled bool) Option {
	return func(l *Logger) { l.enabled = enabled }
}

// New creates an audit Logger. Audit logging is disabled by default.
func New(opts ...Option) *Logger {
	l := &Logger{w: os.Stderr, enabled: false}
	for _, o := range opts {
		o(l)
	}
	return l
}

// Log writes an audit event. It is a no-op when the logger is disabled.
func (l *Logger) Log(e Event) error {
	if !l.enabled {
		return nil
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = l.w.Write(b)
	return err
}
