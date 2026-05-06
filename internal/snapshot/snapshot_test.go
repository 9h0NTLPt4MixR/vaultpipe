package snapshot_test

import (
	"testing"

	"github.com/yourusername/vaultpipe/internal/snapshot"
)

func TestNew_EmptyMap(t *testing.T) {
	s, err := snapshot.New(map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Digest() == "" {
		t.Error("expected non-empty digest")
	}
}

func TestNew_NilMap(t *testing.T) {
	s, err := snapshot.New(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Digest() == "" {
		t.Error("expected non-empty digest for nil input")
	}
}

func TestDigest_DeterministicAcrossKeyOrder(t *testing.T) {
	a, _ := snapshot.New(map[string]string{"foo": "1", "bar": "2"})
	b, _ := snapshot.New(map[string]string{"bar": "2", "foo": "1"})

	if a.Digest() != b.Digest() {
		t.Errorf("digests differ for same content: %s vs %s", a.Digest(), b.Digest())
	}
}

func TestEqual_SameContent(t *testing.T) {
	a, _ := snapshot.New(map[string]string{"key": "value"})
	b, _ := snapshot.New(map[string]string{"key": "value"})

	if !a.Equal(b) {
		t.Error("expected snapshots with same content to be equal")
	}
}

func TestEqual_DifferentContent(t *testing.T) {
	a, _ := snapshot.New(map[string]string{"key": "old"})
	b, _ := snapshot.New(map[string]string{"key": "new"})

	if a.Equal(b) {
		t.Error("expected snapshots with different content to be unequal")
	}
}

func TestEqual_NilOther(t *testing.T) {
	a, _ := snapshot.New(map[string]string{"key": "value"})
	if a.Equal(nil) {
		t.Error("expected Equal(nil) to return false")
	}
}

func TestDiff_DetectsChangedKey(t *testing.T) {
	a, _ := snapshot.New(map[string]string{"foo": "1", "bar": "2"})
	b, _ := snapshot.New(map[string]string{"foo": "1", "bar": "changed"})

	diff := a.Diff(b)
	if len(diff) != 1 || diff[0] != "bar" {
		t.Errorf("expected diff [bar], got %v", diff)
	}
}

func TestDiff_DetectsAddedKey(t *testing.T) {
	a, _ := snapshot.New(map[string]string{"foo": "1"})
	b, _ := snapshot.New(map[string]string{"foo": "1", "bar": "new"})

	diff := a.Diff(b)
	if len(diff) != 1 || diff[0] != "bar" {
		t.Errorf("expected diff [bar], got %v", diff)
	}
}

func TestDiff_DetectsRemovedKey(t *testing.T) {
	a, _ := snapshot.New(map[string]string{"foo": "1", "gone": "bye"})
	b, _ := snapshot.New(map[string]string{"foo": "1"})

	diff := a.Diff(b)
	if len(diff) != 1 || diff[0] != "gone" {
		t.Errorf("expected diff [gone], got %v", diff)
	}
}

func TestDiff_NoDiff_ReturnsEmpty(t *testing.T) {
	a, _ := snapshot.New(map[string]string{"x": "1"})
	b, _ := snapshot.New(map[string]string{"x": "1"})

	if diff := a.Diff(b); len(diff) != 0 {
		t.Errorf("expected no diff, got %v", diff)
	}
}

func TestData_ReturnsCopy(t *testing.T) {
	orig := map[string]string{"a": "1"}
	s, _ := snapshot.New(orig)

	d := s.Data()
	d["a"] = "mutated"

	if s.Data()["a"] != "1" {
		t.Error("expected Data() to return an isolated copy")
	}
}
