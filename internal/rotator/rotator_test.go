package rotator_test

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/vaultpipe/internal/logger"
	"github.com/your-org/vaultpipe/internal/rotator"
)

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	l, _ := logger.New(logger.WithWriter(io.Discard))
	return l
}

func TestNew_Defaults(t *testing.T) {
	r := rotator.New(newTestLogger(t))
	if r == nil {
		t.Fatal("expected non-nil Rotator")
	}
	if !r.LastRotation().IsZero() {
		t.Error("expected zero LastRotation before any rotation")
	}
}

func TestWithInterval_Applied(t *testing.T) {
	r := rotator.New(newTestLogger(t), rotator.WithInterval(10*time.Second))
	if r == nil {
		t.Fatal("expected non-nil Rotator")
	}
}

func TestWithThreshold_Applied(t *testing.T) {
	r := rotator.New(newTestLogger(t), rotator.WithThreshold(2*time.Minute))
	if r == nil {
		t.Fatal("expected non-nil Rotator")
	}
}

func TestRun_TriggersRotationWhenPastThreshold(t *testing.T) {
	log := newTestLogger(t)
	r := rotator.New(log,
		rotator.WithInterval(20*time.Millisecond),
		rotator.WithThreshold(10*time.Minute),
	)

	// Expiry is in the past, so threshold is already exceeded.
	expiry := time.Now().Add(-1 * time.Hour)

	var called atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	_ = r.Run(ctx, expiry, func(_ context.Context) error {
		called.Add(1)
		return nil
	})

	if called.Load() == 0 {
		t.Error("expected RotateFunc to be called at least once")
	}
	if r.LastRotation().IsZero() {
		t.Error("expected LastRotation to be set after successful rotation")
	}
}

func TestRun_DoesNotRotateWhenFarFromExpiry(t *testing.T) {
	log := newTestLogger(t)
	r := rotator.New(log,
		rotator.WithInterval(20*time.Millisecond),
		rotator.WithThreshold(5*time.Minute),
	)

	// Expiry is far in the future.
	expiry := time.Now().Add(2 * time.Hour)

	var called atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = r.Run(ctx, expiry, func(_ context.Context) error {
		called.Add(1)
		return nil
	})

	if called.Load() != 0 {
		t.Errorf("expected RotateFunc not to be called, got %d calls", called.Load())
	}
}

func TestRun_ContinuesAfterRotationError(t *testing.T) {
	log := newTestLogger(t)
	r := rotator.New(log,
		rotator.WithInterval(20*time.Millisecond),
		rotator.WithThreshold(10*time.Minute),
	)

	expiry := time.Now().Add(-1 * time.Hour)

	var called atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	_ = r.Run(ctx, expiry, func(_ context.Context) error {
		called.Add(1)
		return errors.New("vault unavailable")
	})

	// Should have been attempted multiple times despite errors.
	if called.Load() < 2 {
		t.Errorf("expected multiple attempts, got %d", called.Load())
	}
	if !r.LastRotation().IsZero() {
		t.Error("expected LastRotation to remain zero after failed rotations")
	}
}

func TestRun_ReturnsOnContextCancel(t *testing.T) {
	log := newTestLogger(t)
	r := rotator.New(log, rotator.WithInterval(50*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := r.Run(ctx, time.Now().Add(1*time.Hour), func(_ context.Context) error {
		return nil
	})
	if err == nil {
		t.Error("expected error on cancelled context")
	}
}
