// Package template provides lightweight path templating for vaultpipe.
//
// Vault secret paths can be parameterised using Go's text/template syntax.
// Two data sources are available inside every template:
//
//   - .Env  — a snapshot of the process environment at render time.
//   - .Var  — an optional map of extra key/value pairs supplied via WithVars.
//
// Example path template:
//
//	"secret/data/{{ .Env.APP_ENV }}/database"
//
// The Renderer is safe for concurrent use once constructed.
package template
