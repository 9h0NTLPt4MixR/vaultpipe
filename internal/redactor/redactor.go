// Package redactor provides utilities for masking sensitive secret values
// in log output, error messages, and audit trails.
package redactor

import "strings"

const mask = "[REDACTED]"

// Redactor replaces known secret values with a fixed mask string.
type Redactor struct {
	secrets []string
}

// Option is a functional option for Redactor.
type Option func(*Redactor)

// WithSecrets seeds the redactor with an initial set of values to mask.
func WithSecrets(values []string) Option {
	return func(r *Redactor) {
		for _, v := range values {
			if v != "" {
				r.secrets = append(r.secrets, v)
			}
		}
	}
}

// New creates a new Redactor, optionally pre-loaded with secret values.
func New(opts ...Option) *Redactor {
	r := &Redactor{}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Add registers additional secret values to be masked.
func (r *Redactor) Add(values ...string) {
	for _, v := range values {
		if v != "" {
			r.secrets = append(r.secrets, v)
		}
	}
}

// Redact replaces all registered secret values found in s with the mask.
func (r *Redactor) Redact(s string) string {
	for _, secret := range r.secrets {
		s = strings.ReplaceAll(s, secret, mask)
	}
	return s
}

// RedactMap returns a copy of m where all values have been redacted.
// Keys are preserved unchanged.
func (r *Redactor) RedactMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = r.Redact(v)
	}
	return out
}
