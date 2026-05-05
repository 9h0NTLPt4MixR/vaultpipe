package watcher

import (
	"encoding/json"
	"io"
	"time"
)

// ChangeEvent records a single secret-rotation event for downstream consumers
// such as the audit log or a webhook notifier.
type ChangeEvent struct {
	Timestamp time.Time         `json:"timestamp"`
	Path      string            `json:"path"`
	Keys      []string          `json:"keys"`
}

// Notifier wraps a ChangeHandler and emits a JSON ChangeEvent to a writer
// for every change detected, in addition to calling the inner handler.
type Notifier struct {
	inner  ChangeHandler
	writer io.Writer
}

// NewNotifier returns a ChangeHandler that writes JSON change events to w and
// forwards the call to inner.
func NewNotifier(inner ChangeHandler, w io.Writer) ChangeHandler {
	n := &Notifier{inner: inner, writer: w}
	return n.handle
}

func (n *Notifier) handle(path string, secrets map[string]string) {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	event := ChangeEvent{
		Timestamp: time.Now().UTC(),
		Path:      path,
		Keys:      keys,
	}
	data, err := json.Marshal(event)
	if err == nil {
		_, _ = n.writer.Write(append(data, '\n'))
	}
	if n.inner != nil {
		n.inner(path, secrets)
	}
}
