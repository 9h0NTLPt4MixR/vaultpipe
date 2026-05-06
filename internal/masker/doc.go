// Package masker provides an io.Writer wrapper that intercepts byte
// streams and replaces any registered secret substrings with the fixed
// placeholder string "***REDACTED***" before the data reaches the
// underlying writer.
//
// It is intended to be composed with loggers, audit writers, or any
// other output sink that might inadvertently include secret material
// sourced from HashiCorp Vault.
//
// Usage:
//
//	var buf bytes.Buffer
//	m := masker.New(&buf, masker.WithSecrets(secretValues))
//	logger := logger.New(masker.New(os.Stderr, masker.WithSecrets(secrets)))
//
// Additional secrets discovered at runtime can be registered via Add.
package masker
