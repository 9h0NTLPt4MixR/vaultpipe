package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/vaultpipe/internal/ratelimit"
)

func TestNew_Defaults(t *testing.T) {
	l, err := ratelimit.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}
}

func TestNew_InvalidRate(t *testing.T) {
	_, err := ratelimit.New(ratelimit.WithRate(-1))
	if err == nil {
		t.Fatal("expected error for negative rate")
	}
}

func TestNew_InvalidBurst(t *testing.T) {
	_, err := ratelimit.New(ratelimit.WithBurst(0))
	if err == nil {
		t.Fatal("expected error for zero burst")
	}
}

func TestTryAcquire_FirstCallSucceeds(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.WithRate(1), ratelimit.WithBurst(1))
	if !l.TryAcquire() {
		t.Fatal("expected first TryAcquire to succeed")
	}
}

func TestTryAcquire_SecondCallFailsWhenEmpty(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.WithRate(1), ratelimit.WithBurst(1))
	l.TryAcquire() // consume the single token
	if l.TryAcquire() {
		t.Fatal("expected second TryAcquire to fail when bucket is empty")
	}
}

func TestTryAcquire_BurstAllowsMultiple(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.WithRate(1), ratelimit.WithBurst(3))
	for i := 0; i < 3; i++ {
		if !l.TryAcquire() {
			t.Fatalf("expected TryAcquire %d to succeed with burst=3", i+1)
		}
	}
	if l.TryAcquire() {
		t.Fatal("expected 4th TryAcquire to fail")
	}
}

func TestWait_AcquiresImmediatelyWhenTokenAvailable(t *testing.T) {
	l, _ := ratelimit.New(ratelimit.WithRate(10), ratelimit.WithBurst(1))
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWait_RespectsContextCancellation(t *testing.T) {
	// Rate so slow a token will never arrive in time.
	l, _ := ratelimit.New(ratelimit.WithRate(0.001), ratelimit.WithBurst(1))
	l.TryAcquire() // drain the initial token

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := l.Wait(ctx)
	if err == nil {
		t.Fatal("expected context deadline error")
	}
}

func TestWait_ReplenishesTokensOverTime(t *testing.T) {
	// 10 tokens/sec means a token every 100 ms.
	l, _ := ratelimit.New(ratelimit.WithRate(10), ratelimit.WithBurst(1))
	l.TryAcquire() // drain

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := l.Wait(ctx); err != nil {
		t.Fatalf("expected token to be replenished within 500ms: %v", err)
	}
}
