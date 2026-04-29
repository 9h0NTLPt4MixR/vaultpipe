package injector

import (
	"bytes"
	"strings"
	"testing"
)

func TestEnvWriter_SimpleValues(t *testing.T) {
	var buf bytes.Buffer
	ew := NewEnvWriter(&buf)

	err := ew.Write(map[string]string{
		"FOO": "bar",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "FOO=bar") {
		t.Errorf("expected FOO=bar in output, got: %q", output)
	}
}

func TestEnvWriter_QuotesValueWithSpaces(t *testing.T) {
	var buf bytes.Buffer
	ew := NewEnvWriter(&buf)

	err := ew.Write(map[string]string{
		"GREETING": "hello world",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `GREETING="hello world"`) {
		t.Errorf("expected quoted value in output, got: %q", output)
	}
}

func TestEnvWriter_MultipleKeys(t *testing.T) {
	var buf bytes.Buffer
	ew := NewEnvWriter(&buf)

	err := ew.Write(map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "KEY1=value1") {
		t.Errorf("missing KEY1=value1 in output: %q", output)
	}
	if !strings.Contains(output, "KEY2=value2") {
		t.Errorf("missing KEY2=value2 in output: %q", output)
	}
}

func TestQuoteIfNeeded(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with space", `"with space"`},
		{"with=equals", `"with=equals"`},
		{"with\nnewline", `"with\nnewline"`},
		{`has"quote`, `"has\"quote"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := quoteIfNeeded(tt.input)
			if got != tt.expected {
				t.Errorf("quoteIfNeeded(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
