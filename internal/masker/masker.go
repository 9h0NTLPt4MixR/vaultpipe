// Package masker provides utilities for masking sensitive values in
// structured log output and diagnostic strings. It wraps an io.Writer
// and replaces any registered secret substrings with a fixed placeholder
// before the bytes reach the underlying writer.
package masker

import (
	"io"
	"strings"
	"sync"
)

const placeholder = "***REDACTED***"

// Masker wraps an io.Writer and replaces registered secret values with
// a fixed placeholder on every Write call.
type Masker struct {
	mu      sync.RWMutex
	secrets []string
	w       io.Writer
}

// Option is a functional option for Masker.
type Option func(*Masker)

// WithSecrets pre-registers a slice of secret values to mask.
func WithSecrets(secrets []string) Option {
	return func(m *Masker) {
		for _, s := range secrets {
			if s != "" {
				m.secrets = append(m.secrets, s)
			}
		}
	}
}

// New creates a new Masker that writes to w.
func New(w io.Writer, opts ...Option) *Masker {
	m := &Masker{w: w}
	for _, o := range opts {
		o(m)
	}
	return m
}

// Add registers an additional secret value at runtime.
func (m *Masker) Add(secret string) {
	if secret == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.secrets = append(m.secrets, secret)
}

// Write implements io.Writer. It replaces all registered secrets in p
// before forwarding the (possibly modified) bytes to the underlying writer.
func (m *Masker) Write(p []byte) (int, error) {
	m.mu.RLock()
	line := string(p)
	for _, s := range m.secrets {
		line = strings.ReplaceAll(line, s, placeholder)
	}
	m.mu.RUnlock()

	_, err := io.WriteString(m.w, line)
	if err != nil {
		return 0, err
	}
	// Report the original length so callers are not confused.
	return len(p), nil
}
