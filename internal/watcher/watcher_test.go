package watcher_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/vaultpipe/internal/logger"
	"github.com/yourorg/vaultpipe/internal/watcher"
)

// stubFetcher is a thread-safe fake SecretFetcher.
type stubFetcher struct {
	mu      sync.Mutex
	results map[string]map[string]string
	err     error
	calls   int
}

func (s *stubFetcher) GetSecrets(_ context.Context, path string) (map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.results[path], nil
}

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	l, _ := logger.New("error")
	return l
}

func TestNew_Defaults(t *testing.T) {
	f := &stubFetcher{}
	w := watcher.New(f, func(_ string, _ map[string]string) {}, newTestLogger(t))
	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
}

func TestWatcher_CallsHandlerOnFirstPoll(t *testing.T) {
	f := &stubFetcher{
		results: map[string]map[string]string{
			"secret/app": {"KEY": "val"},
		},
	}
	var mu sync.Mutex
	var called []string

	w := watcher.New(f, func(path string, _ map[string]string) {
		mu.Lock()
		called = append(called, path)
		mu.Unlock()
	}, newTestLogger(t),
		watcher.WithPaths("secret/app"),
		watcher.WithInterval(500*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	mu.Lock()
	defer mu.Unlock()
	if len(called) == 0 {
		t.Fatal("expected handler to be called at least once")
	}
	if called[0] != "secret/app" {
		t.Errorf("got path %q, want %q", called[0], "secret/app")
	}
}

func TestWatcher_NoHandlerOnUnchangedSecrets(t *testing.T) {
	f := &stubFetcher{
		results: map[string]map[string]string{
			"secret/stable": {"KEY": "same"},
		},
	}
	var mu sync.Mutex
	callCount := 0

	w := watcher.New(f, func(_ string, _ map[string]string) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}, newTestLogger(t),
		watcher.WithPaths("secret/stable"),
		watcher.WithInterval(50*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	mu.Lock()
	defer mu.Unlock()
	// First poll triggers handler; subsequent polls with same data should not.
	if callCount > 1 {
		t.Errorf("handler called %d times for unchanged secrets, want 1", callCount)
	}
}

func TestWatcher_FetchErrorDoesNotPanic(t *testing.T) {
	f := &stubFetcher{err: errors.New("vault unavailable")}

	w := watcher.New(f, func(_ string, _ map[string]string) {
		t.Error("handler must not be called on fetch error")
	}, newTestLogger(t),
		watcher.WithPaths("secret/app"),
		watcher.WithInterval(500*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	w.Run(ctx) // must not panic
}

func TestWithInterval_ZeroIgnored(t *testing.T) {
	f := &stubFetcher{results: map[string]map[string]string{}}
	// Should not panic; zero interval is ignored, default 30s used.
	w := watcher.New(f, func(_ string, _ map[string]string) {}, newTestLogger(t),
		watcher.WithInterval(0),
	)
	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
}
