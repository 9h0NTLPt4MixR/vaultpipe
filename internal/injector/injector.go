package injector

import (
	"fmt"
	"os"
	"strings"
)

// Injector handles injecting secrets into process environments.
type Injector struct {
	prefix    string
	overwrite bool
}

// Option is a functional option for configuring an Injector.
type Option func(*Injector)

// WithPrefix sets a prefix to apply to all injected environment variable names.
func WithPrefix(prefix string) Option {
	return func(i *Injector) {
		i.prefix = strings.ToUpper(prefix)
	}
}

// WithOverwrite allows the injector to overwrite existing environment variables.
func WithOverwrite(overwrite bool) Option {
	return func(i *Injector) {
		i.overwrite = overwrite
	}
}

// NewInjector creates a new Injector with the provided options.
func NewInjector(opts ...Option) *Injector {
	i := &Injector{}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

// Inject takes a map of secret key/value pairs and sets them as environment variables.
// Keys are uppercased and optionally prefixed.
func (i *Injector) Inject(secrets map[string]string) error {
	for key, value := range secrets {
		envKey := i.formatKey(key)
		if !i.overwrite {
			if existing := os.Getenv(envKey); existing != "" {
				continue
			}
		}
		if err := os.Setenv(envKey, value); err != nil {
			return fmt.Errorf("injector: failed to set env var %q: %w", envKey, err)
		}
	}
	return nil
}

// InjectToMap returns a new map with formatted keys suitable for use as env vars,
// without modifying the actual process environment.
func (i *Injector) InjectToMap(secrets map[string]string) map[string]string {
	result := make(map[string]string, len(secrets))
	for key, value := range secrets {
		result[i.formatKey(key)] = value
	}
	return result
}

// formatKey uppercases the key and prepends the prefix if set.
func (i *Injector) formatKey(key string) string {
	upper := strings.ToUpper(key)
	if i.prefix != "" {
		return i.prefix + "_" + upper
	}
	return upper
}
