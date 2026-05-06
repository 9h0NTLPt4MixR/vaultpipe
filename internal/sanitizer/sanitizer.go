// Package sanitizer provides utilities for validating and normalising
// secret key names before they are injected into process environments.
// Keys that contain invalid characters are either rejected or rewritten
// according to the configured policy.
package sanitizer

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// validKey matches POSIX-portable environment variable names.
	validKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	// invalidChars matches any character that is not allowed in an env key.
	invalidChars = regexp.MustCompile(`[^A-Za-z0-9_]`)

	// ErrEmptyKey is returned when an empty key is supplied.
	ErrEmptyKey = errors.New("sanitizer: key must not be empty")
	// ErrInvalidKey is returned when strict mode is enabled and the key
	// contains characters that are not allowed.
	ErrInvalidKey = errors.New("sanitizer: key contains invalid characters")
)

// Option is a functional option for Sanitizer.
type Option func(*Sanitizer)

// WithStrict enables strict mode: keys that do not conform to the
// POSIX portable character set are rejected rather than rewritten.
func WithStrict(strict bool) Option {
	return func(s *Sanitizer) { s.strict = strict }
}

// WithUppercase causes all sanitised keys to be converted to upper-case,
// which is the conventional form for environment variable names.
func WithUppercase(upper bool) Option {
	return func(s *Sanitizer) { s.uppercase = upper }
}

// Sanitizer validates and normalises secret key names.
type Sanitizer struct {
	strict    bool
	uppercase bool
}

// New returns a Sanitizer configured with the supplied options.
// By default the sanitizer rewrites invalid characters to underscores
// and converts keys to upper-case.
func New(opts ...Option) *Sanitizer {
	s := &Sanitizer{uppercase: true}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Sanitize validates and normalises key. It returns the sanitised key or
// an error when strict mode is enabled and the key is non-conforming.
func (s *Sanitizer) Sanitize(key string) (string, error) {
	if key == "" {
		return "", ErrEmptyKey
	}

	if s.strict && !validKey.MatchString(key) {
		return "", ErrInvalidKey
	}

	// Replace invalid characters with underscores.
	key = invalidChars.ReplaceAllString(key, "_")

	// Ensure the key does not start with a digit after replacement.
	if len(key) > 0 && key[0] >= '0' && key[0] <= '9' {
		key = "_" + key
	}

	if s.uppercase {
		key = strings.ToUpper(key)
	}

	return key, nil
}

// SanitizeMap applies Sanitize to every key in m and returns a new map.
// Collisions caused by normalisation are resolved by keeping the last
// value in iteration order. An error is returned immediately if any key
// fails validation in strict mode.
func (s *Sanitizer) SanitizeMap(m map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(m))
	for k, v := range m {
		clean, err := s.Sanitize(k)
		if err != nil {
			return nil, err
		}
		out[clean] = v
	}
	return out, nil
}
