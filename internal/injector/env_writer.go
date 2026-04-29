package injector

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// EnvWriter serializes a map of environment variables into .env file format.
type EnvWriter struct {
	w io.Writer
}

// NewEnvWriter creates a new EnvWriter that writes to the given io.Writer.
func NewEnvWriter(w io.Writer) *EnvWriter {
	return &EnvWriter{w: w}
}

// Write serializes the given key/value pairs as KEY=VALUE lines.
// Values containing spaces or special characters are quoted.
func (e *EnvWriter) Write(env map[string]string) error {
	bw := bufio.NewWriter(e.w)
	for key, value := range env {
		line := fmt.Sprintf("%s=%s\n", key, quoteIfNeeded(value))
		if _, err := bw.WriteString(line); err != nil {
			return fmt.Errorf("env_writer: failed to write key %q: %w", key, err)
		}
	}
	return bw.Flush()
}

// quoteIfNeeded wraps the value in double quotes if it contains
// spaces, equals signs, or newline characters.
func quoteIfNeeded(value string) string {
	if strings.ContainsAny(value, " \t\n=") {
		escaped := strings.ReplaceAll(value, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return value
}
