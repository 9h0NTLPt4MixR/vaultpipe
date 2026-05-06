package backoff_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpipe/internal/backoff"
)

func TestDefaultConfig_SaneValues(t *testing.T) {
	cfg := backoff.DefaultConfig()
	if cfg.BaseDelay <= 0 {
		t.Fatalf("expected positive BaseDelay, got %v", cfg.BaseDelay)
	}
	if cfg.MaxDelay < cfg.BaseDelay {
		t.Fatalf("MaxDelay %v < BaseDelay %v", cfg.MaxDelay, cfg.BaseDelay)
	}
	if cfg.Multiplier <= 1 {
		t.Fatalf("expected Multiplier > 1, got %v", cfg.Multiplier)
	}
}

func TestNext_IncreasesWithAttempt(t *testing.T) {
	s := backoff.New(backoff.Config{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		Multiplier: 2.0,
		Jitter:     false,
	})

	prev := s.Next(0)
	for i := 1; i <= 5; i++ {
		curr := s.Next(i)
		if curr <= prev {
			t.Fatalf("attempt %d: expected delay %v > %v", i, curr, prev)
		}
		prev = curr
	}
}

func TestNext_CapsAtMaxDelay(t *testing.T) {
	max := 500 * time.Millisecond
	s := backoff.New(backoff.Config{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   max,
		Multiplier: 2.0,
		Jitter:     false,
	})

	for i := 0; i <= 20; i++ {
		d := s.Next(i)
		if d > max {
			t.Fatalf("attempt %d: delay %v exceeds max %v", i, d, max)
		}
	}
}

func TestNext_JitterDoesNotExceedMax(t *testing.T) {
	max := 1 * time.Second
	s := backoff.New(backoff.Config{
		BaseDelay:  50 * time.Millisecond,
		MaxDelay:   max,
		Multiplier: 2.0,
		Jitter:     true,
	})

	for i := 0; i < 100; i++ {
		d := s.Next(i % 10)
		if d > max {
			t.Fatalf("jittered delay %v exceeds max %v", d, max)
		}
	}
}

func TestNew_InvalidMultiplierDefaultsToTwo(t *testing.T) {
	s := backoff.New(backoff.Config{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 0.5, // invalid
		Jitter:     false,
	})

	d0 := s.Next(0)
	d1 := s.Next(1)
	if d1 != d0*2 {
		t.Fatalf("expected multiplier=2 fallback: got d0=%v d1=%v", d0, d1)
	}
}
