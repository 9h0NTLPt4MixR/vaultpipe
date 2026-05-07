// Package resolver provides path resolution for Vault secret paths,
// supporting both static paths and environment-variable-driven dynamic paths.
package resolver

import (
	"fmt"
	"os"
	"strings"
)

// Resolver resolves a raw Vault secret path, expanding any ${ENV_VAR}
// placeholders found within it.
type Resolver struct {
	envLookup func(string) (string, bool)
}

// Option is a functional option for Resolver.
type Option func(*Resolver)

// WithEnvLookup replaces the default os.LookupEnv with a custom function.
// Primarily useful for testing.
func WithEnvLookup(fn func(string) (string, bool)) Option {
	return func(r *Resolver) {
		r.envLookup = fn
	}
}

// New returns a new Resolver with the provided options applied.
func New(opts ...Option) *Resolver {
	r := &Resolver{
		envLookup: os.LookupEnv,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Resolve expands ${VAR} placeholders in path using the configured env lookup.
// Returns an error if any referenced variable is not set.
func (r *Resolver) Resolve(path string) (string, error) {
	var missing []string

	resolved := os.Expand(path, func(key string) string {
		val, ok := r.envLookup(key)
		if !ok {
			missing = append(missing, key)
			return ""
		}
		return val
	})

	if len(missing) > 0 {
		return "", fmt.Errorf("resolver: unresolved environment variables: %s",
			strings.Join(missing, ", "))
	}

	return resolved, nil
}

// ResolveAll resolves each path in the provided slice, returning the fully
// expanded slice or the first error encountered.
func (r *Resolver) ResolveAll(paths []string) ([]string, error) {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		resolved, err := r.Resolve(p)
		if err != nil {
			return nil, err
		}
		out = append(out, resolved)
	}
	return out, nil
}
