// Package resolver resolves Vault secret paths that contain ${ENV_VAR}
// placeholder syntax.
//
// At startup vaultpipe may be configured with paths such as:
//
//	"secret/data/${APP_ENV}/database"
//
// The Resolver expands these placeholders using the process environment (or a
// custom lookup function supplied via WithEnvLookup), returning an error if any
// referenced variable is absent.  This keeps secret-path configuration
// portable across deployment environments without requiring separate config
// files per environment.
package resolver
