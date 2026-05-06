package masker_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/vaultpipe/internal/masker"
)

func TestWrite_NoSecrets_PassesThrough(t *testing.T) {
	var buf bytes.Buffer
	m := masker.New(&buf)

	_, err := m.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestWrite_MasksRegisteredSecret(t *testing.T) {
	var buf bytes.Buffer
	m := masker.New(&buf, masker.WithSecrets([]string{"s3cr3t"}))

	m.Write([]byte("token=s3cr3t")) //nolint:errcheck

	if strings.Contains(buf.String(), "s3cr3t") {
		t.Errorf("secret should have been masked, got: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "***REDACTED***") {
		t.Errorf("expected placeholder in output, got: %q", buf.String())
	}
}

func TestWrite_MultipleSecrets(t *testing.T) {
	var buf bytes.Buffer
	m := masker.New(&buf, masker.WithSecrets([]string{"alpha", "beta"}))

	m.Write([]byte("alpha and beta are both secret")) //nolint:errcheck

	result := buf.String()
	if strings.Contains(result, "alpha") || strings.Contains(result, "beta") {
		t.Errorf("secrets should be masked, got: %q", result)
	}
}

func TestAdd_RegistersNewSecret(t *testing.T) {
	var buf bytes.Buffer
	m := masker.New(&buf)
	m.Add("runtime-secret")

	m.Write([]byte("value=runtime-secret")) //nolint:errcheck

	if strings.Contains(buf.String(), "runtime-secret") {
		t.Errorf("dynamically added secret should be masked, got: %q", buf.String())
	}
}

func TestAdd_IgnoresEmptyString(t *testing.T) {
	var buf bytes.Buffer
	m := masker.New(&buf)
	m.Add("") // should not panic or register

	m.Write([]byte("some output")) //nolint:errcheck

	if buf.String() != "some output" {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestWrite_ReturnsOriginalLength(t *testing.T) {
	var buf bytes.Buffer
	m := masker.New(&buf, masker.WithSecrets([]string{"x"}))

	input := []byte("prefix-x-suffix")
	n, err := m.Write(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(input) {
		t.Errorf("expected n=%d, got %d", len(input), n)
	}
}

func TestWithSecrets_IgnoresEmptyEntries(t *testing.T) {
	var buf bytes.Buffer
	m := masker.New(&buf, masker.WithSecrets([]string{"", "", "real"}))

	m.Write([]byte("real value here")) //nolint:errcheck

	if strings.Contains(buf.String(), "real") {
		t.Errorf("non-empty secret should be masked, got: %q", buf.String())
	}
}
