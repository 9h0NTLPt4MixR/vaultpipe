// Package filter provides key-based filtering for secrets maps,
// allowing callers to include or exclude specific keys before injection.
package filter

import "strings"

// Option configures a Filter.
type Option func(*Filter)

// Filter selectively includes or excludes secret keys.
type Filter struct {
	allowlist map[string]struct{}
	denylist  map[string]struct{}
	prefix   string
}

// WithAllowlist restricts the output to only the provided keys.
// An empty allowlist means all keys are permitted (unless denied).
func WithAllowlist(keys ...string) Option {
	return func(f *Filter) {
		for _, k := range keys {
			f.allowlist[k] = struct{}{}
		}
	}
}

// WithDenylist excludes the provided keys from the output.
func WithDenylist(keys ...string) Option {
	return func(f *Filter) {
		for _, k := range keys {
			f.denylist[k] = struct{}{}
		}
	}
}

// WithKeyPrefix retains only keys that begin with the given prefix.
func WithKeyPrefix(prefix string) Option {
	return func(f *Filter) {
		f.prefix = prefix
	}
}

// New creates a Filter with the supplied options.
func New(opts ...Option) *Filter {
	f := &Filter{
		allowlist: make(map[string]struct{}),
		denylist:  make(map[string]struct{}),
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Apply returns a new map containing only the entries that pass the filter.
func (f *Filter) Apply(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if !f.permitted(k) {
			continue
		}
		out[k] = v
	}
	return out
}

func (f *Filter) permitted(key string) bool {
	if _, denied := f.denylist[key]; denied {
		return false
	}
	if f.prefix != "" && !strings.HasPrefix(key, f.prefix) {
		return false
	}
	if len(f.allowlist) > 0 {
		_, ok := f.allowlist[key]
		return ok
	}
	return true
}
