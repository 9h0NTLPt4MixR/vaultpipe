// Package template provides secret path templating for vaultpipe,
// allowing dynamic Vault paths to be constructed from environment variables
// or runtime values.
package template

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// Renderer renders Vault secret path templates using environment variables
// and a supplied data map as the template context.
type Renderer struct {
	extraVars map[string]string
}

// Option configures a Renderer.
type Option func(*Renderer)

// WithVars supplies additional key/value pairs available inside templates
// alongside environment variables.
func WithVars(vars map[string]string) Option {
	return func(r *Renderer) {
		for k, v := range vars {
			r.extraVars[k] = v
		}
	}
}

// New returns a Renderer configured with the supplied options.
func New(opts ...Option) *Renderer {
	r := &Renderer{extraVars: make(map[string]string)}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Render executes the given path template and returns the rendered string.
// Templates may reference environment variables via .Env and extra vars via
// .Var, e.g. "secret/{{ .Env.APP_ENV }}/database".
func (r *Renderer) Render(pathTmpl string) (string, error) {
	tmpl, err := template.New("path").Option("missingkey=error").Parse(pathTmpl)
	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}

	data := map[string]interface{}{
		"Env": envMap(),
		"Var": r.extraVars,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}
	return buf.String(), nil
}

// envMap returns a snapshot of the current process environment as a map.
func envMap() map[string]string {
	m := make(map[string]string)
	for _, kv := range os.Environ() {
		for i := 0; i < len(kv); i++ {
			if kv[i] == '=' {
				m[kv[:i]] = kv[i+1:]
				break
			}
		}
	}
	return m
}
