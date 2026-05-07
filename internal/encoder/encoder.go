// Package encoder provides utilities for encoding and decoding secret
// values between different representations (e.g. base64, hex) before
// injection into process environments.
package encoder

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// Encoding identifies the encoding scheme to apply.
type Encoding int

const (
	// Raw leaves the value unchanged.
	Raw Encoding = iota
	// Base64Std encodes using standard base64 (RFC 4648).
	Base64Std
	// Base64URL encodes using URL-safe base64 (RFC 4648 §5).
	Base64URL
	// Hex encodes using lowercase hexadecimal.
	Hex
)

// Encoder transforms secret values according to a chosen Encoding.
type Encoder struct {
	enc Encoding
}

// Option configures an Encoder.
type Option func(*Encoder)

// WithEncoding sets the encoding scheme.
func WithEncoding(e Encoding) Option {
	return func(enc *Encoder) {
		enc.enc = e
	}
}

// New returns a new Encoder. The default encoding is Raw (no transformation).
func New(opts ...Option) *Encoder {
	e := &Encoder{enc: Raw}
	for _, o := range opts {
		o(e)
	}
	return e
}

// Encode transforms value according to the configured encoding scheme.
func (e *Encoder) Encode(value string) (string, error) {
	switch e.enc {
	case Raw:
		return value, nil
	case Base64Std:
		return base64.StdEncoding.EncodeToString([]byte(value)), nil
	case Base64URL:
		return base64.URLEncoding.EncodeToString([]byte(value)), nil
	case Hex:
		return hex.EncodeToString([]byte(value)), nil
	default:
		return "", fmt.Errorf("encoder: unknown encoding %d", e.enc)
	}
}

// EncodeMap applies Encode to every value in secrets, returning a new map.
// The original map is not modified. If any value cannot be encoded the
// function returns a non-nil error and a nil map.
func (e *Encoder) EncodeMap(secrets map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		encoded, err := e.Encode(v)
		if err != nil {
			return nil, fmt.Errorf("encoder: key %q: %w", k, err)
		}
		out[k] = encoded
	}
	return out, nil
}
