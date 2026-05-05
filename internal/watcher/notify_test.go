package watcher_test

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"

	"github.com/yourorg/vaultpipe/internal/watcher"
)

func TestNewNotifier_WritesJSONEvent(t *testing.T) {
	var buf bytes.Buffer
	var innerCalled bool

	inner := func(_ string, _ map[string]string) { innerCalled = true }
	handler := watcher.NewNotifier(inner, &buf)

	secrets := map[string]string{"DB_PASS": "s3cr3t", "API_KEY": "abc"}
	handler("secret/app", secrets)

	if !innerCalled {
		t.Error("inner handler was not called")
	}

	var event watcher.ChangeEvent
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("failed to parse JSON event: %v", err)
	}
	if event.Path != "secret/app" {
		t.Errorf("got path %q, want %q", event.Path, "secret/app")
	}
	if event.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	sort.Strings(event.Keys)
	if len(event.Keys) != 2 {
		t.Errorf("got %d keys, want 2", len(event.Keys))
	}
}

func TestNewNotifier_NilInnerDoesNotPanic(t *testing.T) {
	var buf bytes.Buffer
	handler := watcher.NewNotifier(nil, &buf)
	// Must not panic.
	handler("secret/app", map[string]string{"X": "1"})
	if buf.Len() == 0 {
		t.Error("expected JSON output even with nil inner handler")
	}
}

func TestNewNotifier_EmptySecretsWritesEmptyKeys(t *testing.T) {
	var buf bytes.Buffer
	handler := watcher.NewNotifier(nil, &buf)
	handler("secret/empty", map[string]string{})

	var event watcher.ChangeEvent
	if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(event.Keys) != 0 {
		t.Errorf("expected empty keys slice, got %v", event.Keys)
	}
}
