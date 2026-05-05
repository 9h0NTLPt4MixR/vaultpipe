package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultpipe/internal/audit"
)

func TestNew_Defaults(t *testing.T) {
	l := audit.New()
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestLog_DisabledByDefault(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(audit.WithWriter(&buf))
	err := l.Log(audit.Event{Type: audit.EventSecretFetch, Path: "secret/data/app", Success: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output when disabled, got %q", buf.String())
	}
}

func TestLog_EnabledWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(audit.WithWriter(&buf), audit.WithEnabled(true))

	e := audit.Event{
		Type:    audit.EventSecretFetch,
		Path:    "secret/data/app",
		Keys:    []string{"DB_PASSWORD", "API_KEY"},
		Success: true,
	}
	if err := l.Log(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var got audit.Event
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}
	if got.Type != audit.EventSecretFetch {
		t.Errorf("expected type %q, got %q", audit.EventSecretFetch, got.Type)
	}
	if got.Path != "secret/data/app" {
		t.Errorf("unexpected path: %q", got.Path)
	}
	if len(got.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(got.Keys))
	}
}

func TestLog_SetsTimestampWhenZero(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(audit.WithWriter(&buf), audit.WithEnabled(true))

	before := time.Now().UTC()
	_ = l.Log(audit.Event{Type: audit.EventSecretInject, Success: true})
	after := time.Now().UTC()

	var got audit.Event
	_ = json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got)
	if got.Timestamp.Before(before) || got.Timestamp.After(after) {
		t.Errorf("timestamp %v out of expected range [%v, %v]", got.Timestamp, before, after)
	}
}

func TestLog_ErrorEvent(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(audit.WithWriter(&buf), audit.WithEnabled(true))

	e := audit.Event{
		Type:    audit.EventSecretFetch,
		Path:    "secret/data/missing",
		Success: false,
		Error:   "permission denied",
	}
	_ = l.Log(e)

	var got audit.Event
	_ = json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got)
	if got.Success {
		t.Error("expected success=false")
	}
	if got.Error != "permission denied" {
		t.Errorf("unexpected error field: %q", got.Error)
	}
}
