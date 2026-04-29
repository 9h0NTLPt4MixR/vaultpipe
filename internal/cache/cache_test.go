package cache_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpipe/internal/cache"
)

func TestCache_SetAndGet(t *testing.T) {
	c := cache.New(cache.WithTTL(1 * time.Minute))

	secrets := map[string]string{"KEY": "value"}
	c.Set("secret/app", secrets)

	got, ok := c.Get("secret/app")
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if got["KEY"] != "value" {
		t.Errorf("expected value %q, got %q", "value", got["KEY"])
	}
}

func TestCache_Miss(t *testing.T) {
	c := cache.New()

	_, ok := c.Get("secret/missing")
	if ok {
		t.Fatal("expected cache miss, got hit")
	}
}

func TestCache_Expiry(t *testing.T) {
	now := time.Now()
	nowFunc := func() time.Time { return now }

	c := cache.New(cache.WithTTL(10 * time.Second))
	// Inject controllable clock via unexported field workaround: use a very short TTL
	c2 := cache.New(cache.WithTTL(1 * time.Millisecond))
	_ = nowFunc // used conceptually; c2 relies on real clock + sleep

	c2.Set("secret/app", map[string]string{"K": "v"})
	time.Sleep(5 * time.Millisecond)

	_, ok := c2.Get("secret/app")
	if ok {
		t.Fatal("expected expired cache entry to be a miss")
	}

	// Ensure non-expired entry still works
	c.Set("secret/app", map[string]string{"K": "v"})
	_, ok = c.Get("secret/app")
	if !ok {
		t.Fatal("expected valid cache entry to be a hit")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := cache.New()
	c.Set("secret/app", map[string]string{"A": "1"})
	c.Invalidate("secret/app")

	_, ok := c.Get("secret/app")
	if ok {
		t.Fatal("expected invalidated entry to be a miss")
	}
}

func TestCache_Flush(t *testing.T) {
	c := cache.New()
	c.Set("secret/a", map[string]string{"X": "1"})
	c.Set("secret/b", map[string]string{"Y": "2"})
	c.Flush()

	_, okA := c.Get("secret/a")
	_, okB := c.Get("secret/b")
	if okA || okB {
		t.Fatal("expected all entries to be flushed")
	}
}

func TestCache_IsolatesMutations(t *testing.T) {
	c := cache.New()
	orig := map[string]string{"KEY": "original"}
	c.Set("secret/app", orig)

	// Mutate the original map after storing
	orig["KEY"] = "mutated"

	got, ok := c.Get("secret/app")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got["KEY"] != "original" {
		t.Errorf("cache should store a copy; expected %q, got %q", "original", got["KEY"])
	}
}
