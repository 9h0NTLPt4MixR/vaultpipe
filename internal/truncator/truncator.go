// Package truncator provides utilities for truncating secret values in
// log output and diagnostic messages to avoid accidental exposure of
// sensitive data while still providing enough context for debugging.
package truncator

import "strings"

const (
	// DefaultMaxLen is the default maximum number of visible characters
	// before truncation is applied.
	DefaultMaxLen = 6

	// DefaultSuffix is appended to truncated values.
	DefaultSuffix = "…"

	// maskFull is used when the value is too short to show any characters.
	maskFull = "[redacted]"
)

// Truncator shortens secret values to a safe preview length.
type Truncator struct {
	maxLen int
	suffix string
}

// Option configures a Truncator.
type Option func(*Truncator)

// WithMaxLen sets the maximum number of plaintext characters to retain.
// Panics if n < 1.
func WithMaxLen(n int) Option {
	if n < 1 {
		panic("truncator: maxLen must be >= 1")
	}
	return func(t *Truncator) { t.maxLen = n }
}

// WithSuffix overrides the default truncation suffix.
func WithSuffix(s string) Option {
	return func(t *Truncator) { t.suffix = s }
}

// New returns a Truncator configured with the supplied options.
func New(opts ...Option) *Truncator {
	t := &Truncator{
		maxLen: DefaultMaxLen,
		suffix: DefaultSuffix,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Truncate returns a safe preview of v. Values shorter than or equal to
// maxLen are returned unchanged. Values that are entirely whitespace or
// empty are replaced with the full-mask sentinel.
func (t *Truncator) Truncate(v string) string {
	if strings.TrimSpace(v) == "" {
		return maskFull
	}
	runes := []rune(v)
	if len(runes) <= t.maxLen {
		return v
	}
	return string(runes[:t.maxLen]) + t.suffix
}

// TruncateMap applies Truncate to every value in m and returns a new map.
func (t *Truncator) TruncateMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = t.Truncate(v)
	}
	return out
}
