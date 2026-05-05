package audit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/yourusername/vaultpipe/internal/audit"
)

// stubFetcher is a minimal SecretFetcher for testing.
type stubFetcher struct {
	secrets map[string]string
	err     error
}

func (s *stubFetcher) GetSecrets(_ string) (map[string]string, error) {
	return s.secrets, s.err
}

func TestInstrumentedFetcher_Success(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(audit.WithWriter(&buf), audit.WithEnabled(true))

	stub := &stubFetcher{secrets: map[string]string{"FOO": "bar", "BAZ": "qux"}}
	f := audit.NewInstrumentedFetcher(stub, l)

	got, err := f.GetSecrets("secret/data/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(got))
	}

	var e audit.Event
	_ = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &e)
	if e.Type != audit.EventSecretFetch {
		t.Errorf("expected event type %q, got %q", audit.EventSecretFetch, e.Type)
	}
	if !e.Success {
		t.Error("expected success=true")
	}
	if len(e.Keys) != 2 {
		t.Errorf("expected 2 keys in event, got %d", len(e.Keys))
	}
}

func TestInstrumentedFetcher_Error(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(audit.WithWriter(&buf), audit.WithEnabled(true))

	stub := &stubFetcher{err: errors.New("vault unreachable")}
	f := audit.NewInstrumentedFetcher(stub, l)

	_, err := f.GetSecrets("secret/data/app")
	if err == nil {
		t.Fatal("expected error from fetcher")
	}

	var e audit.Event
	_ = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &e)
	if e.Success {
		t.Error("expected success=false on error")
	}
	if e.Error != "vault unreachable" {
		t.Errorf("unexpected error field: %q", e.Error)
	}
}

func TestInstrumentedFetcher_AuditDisabled(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(audit.WithWriter(&buf)) // disabled by default

	stub := &stubFetcher{secrets: map[string]string{"X": "1"}}
	f := audit.NewInstrumentedFetcher(stub, l)

	_, _ = f.GetSecrets("secret/data/app")
	if buf.Len() != 0 {
		t.Errorf("expected no audit output when disabled, got %q", buf.String())
	}
}
