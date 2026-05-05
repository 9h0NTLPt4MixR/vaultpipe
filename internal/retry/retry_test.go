package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultpipe/internal/retry"
)

var errTransient = errors.New("transient error")

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	calls := 0
	err := retry.Do(context.Background(), retry.DefaultConfig(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnTransientError(t *testing.T) {
	calls := 0
	cfg := retry.Config{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond, Factor: 2}
	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return errTransient
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after eventual success, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsMaxAttempts(t *testing.T) {
	calls := 0
	cfg := retry.Config{MaxAttempts: 4, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, Factor: 2}
	err := retry.Do(context.Background(), cfg, func() error {
		calls++
		return errTransient
	})
	if !errors.Is(err, retry.ErrMaxAttempts) {
		t.Fatalf("expected ErrMaxAttempts, got %v", err)
	}
	if !errors.Is(err, errTransient) {
		t.Fatalf("expected wrapped errTransient, got %v", err)
	}
	if calls != 4 {
		t.Fatalf("expected 4 calls, got %d", calls)
	}
}

func TestDo_ContextCancelledBetweenRetries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	cfg := retry.Config{MaxAttempts: 10, BaseDelay: 50 * time.Millisecond, MaxDelay: time.Second, Factor: 2}
	err := retry.Do(ctx, cfg, func() error {
		calls++
		if calls == 1 {
			cancel()
		}
		return errTransient
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDefaultConfig_SaneValues(t *testing.T) {
	cfg := retry.DefaultConfig()
	if cfg.MaxAttempts <= 0 {
		t.Error("MaxAttempts must be positive")
	}
	if cfg.BaseDelay <= 0 {
		t.Error("BaseDelay must be positive")
	}
	if cfg.MaxDelay < cfg.BaseDelay {
		t.Error("MaxDelay must be >= BaseDelay")
	}
	if cfg.Factor < 1 {
		t.Error("Factor must be >= 1")
	}
}
