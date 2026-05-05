package renewer_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/vaultpipe/internal/logger"
	"github.com/your-org/vaultpipe/internal/renewer"
)

// mockRenewer is a test double for LeaseRenewer.
type mockRenewer struct {
	calls  atomic.Int32
	err    error
	newTTL time.Duration
}

func (m *mockRenewer) RenewLease(_ context.Context, _ string, _ int) (time.Duration, error) {
	m.calls.Add(1)
	return m.newTTL, m.err
}

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	return logger.New("error") // suppress noise in tests
}

func TestNew_Defaults(t *testing.T) {
	mock := &mockRenewer{newTTL: 10 * time.Second}
	r := renewer.New(mock, newTestLogger(t))
	if r == nil {
		t.Fatal("expected non-nil Renewer")
	}
}

func TestWithGraceFraction_Valid(t *testing.T) {
	mock := &mockRenewer{newTTL: 10 * time.Second}
	// Should not panic and renewer should be created.
	r := renewer.New(mock, newTestLogger(t), renewer.WithGraceFraction(0.2))
	if r == nil {
		t.Fatal("expected non-nil Renewer")
	}
}

func TestWithGraceFraction_Invalid(t *testing.T) {
	mock := &mockRenewer{newTTL: 10 * time.Second}
	// Invalid fractions (<=0 or >=1) should be silently ignored; renewer still created.
	r := renewer.New(mock, newTestLogger(t), renewer.WithGraceFraction(1.5))
	if r == nil {
		t.Fatal("expected non-nil Renewer")
	}
}

func TestTrack_CallsRenewBeforeExpiry(t *testing.T) {
	mock := &mockRenewer{newTTL: 50 * time.Millisecond}
	r := renewer.New(mock, newTestLogger(t), renewer.WithGraceFraction(0.5))

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	r.Track(ctx, "lease-abc", 100*time.Millisecond)

	// Wait long enough for at least one renewal.
	time.Sleep(120 * time.Millisecond)

	if mock.calls.Load() == 0 {
		t.Error("expected at least one renewal call")
	}
}

func TestStop_CancelsRenewal(t *testing.T) {
	mock := &mockRenewer{newTTL: 50 * time.Millisecond}
	r := renewer.New(mock, newTestLogger(t), renewer.WithGraceFraction(0.5))

	ctx := context.Background()
	r.Track(ctx, "lease-xyz", 100*time.Millisecond)
	r.Stop("lease-xyz")

	time.Sleep(120 * time.Millisecond)
	callsAfterStop := mock.calls.Load()

	time.Sleep(60 * time.Millisecond)
	if mock.calls.Load() != callsAfterStop {
		t.Error("expected no further renewals after Stop")
	}
}

func TestTrack_RenewError_StopsRetrying(t *testing.T) {
	mock := &mockRenewer{err: errors.New("vault unavailable"), newTTL: 50 * time.Millisecond}
	r := renewer.New(mock, newTestLogger(t), renewer.WithGraceFraction(0.5))

	ctx := context.Background()
	r.Track(ctx, "lease-err", 60*time.Millisecond)
	time.Sleep(150 * time.Millisecond)

	// On error the watcher exits; calls should be exactly 1.
	if mock.calls.Load() != 1 {
		t.Errorf("expected exactly 1 call on error, got %d", mock.calls.Load())
	}
	r.StopAll()
}

func TestStopAll_CancelsAllLeases(t *testing.T) {
	mock := &mockRenewer{newTTL: 200 * time.Millisecond}
	r := renewer.New(mock, newTestLogger(t), renewer.WithGraceFraction(0.5))

	ctx := context.Background()
	r.Track(ctx, "l1", 400*time.Millisecond)
	r.Track(ctx, "l2", 400*time.Millisecond)
	r.StopAll()

	time.Sleep(250 * time.Millisecond)
	// No renewals should have fired because StopAll cancelled before the grace delay.
	if mock.calls.Load() != 0 {
		t.Errorf("expected 0 calls after StopAll, got %d", mock.calls.Load())
	}
}
