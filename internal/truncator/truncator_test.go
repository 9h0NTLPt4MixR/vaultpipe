package truncator

import (
	"strings"
	"testing"
)

func TestTruncate_ShortValueUnchanged(t *testing.T) {
	tr := New()
	got := tr.Truncate("abc")
	if got != "abc" {
		t.Fatalf("expected %q, got %q", "abc", got)
	}
}

func TestTruncate_ExactLengthUnchanged(t *testing.T) {
	tr := New(WithMaxLen(4))
	v := "abcd"
	if got := tr.Truncate(v); got != v {
		t.Fatalf("expected %q, got %q", v, got)
	}
}

func TestTruncate_LongValueTruncated(t *testing.T) {
	tr := New(WithMaxLen(4))
	got := tr.Truncate("supersecret")
	if !strings.HasPrefix(got, "supe") {
		t.Fatalf("expected prefix %q in %q", "supe", got)
	}
	if !strings.HasSuffix(got, DefaultSuffix) {
		t.Fatalf("expected suffix %q in %q", DefaultSuffix, got)
	}
}

func TestTruncate_CustomSuffix(t *testing.T) {
	tr := New(WithMaxLen(3), WithSuffix("..."))
	got := tr.Truncate("password123")
	expected := "pas..."
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestTruncate_EmptyStringMasked(t *testing.T) {
	tr := New()
	if got := tr.Truncate(""); got != maskFull {
		t.Fatalf("expected %q, got %q", maskFull, got)
	}
}

func TestTruncate_WhitespaceOnlyMasked(t *testing.T) {
	tr := New()
	if got := tr.Truncate("   "); got != maskFull {
		t.Fatalf("expected %q, got %q", maskFull, got)
	}
}

func TestTruncate_MultiByte(t *testing.T) {
	tr := New(WithMaxLen(3))
	// Each character is multi-byte (Japanese).
	v := "秘密情報テスト"
	got := tr.Truncate(v)
	runes := []rune(got)
	// Should be 3 runes + suffix rune(s).
	if !strings.HasSuffix(got, DefaultSuffix) {
		t.Fatalf("expected suffix in %q", got)
	}
	visible := []rune(strings.TrimSuffix(got, DefaultSuffix))
	if len(visible) != 3 {
		t.Fatalf("expected 3 visible runes, got %d in %q", len(runes), got)
	}
}

func TestWithMaxLen_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for maxLen=0")
		}
	}()
	WithMaxLen(0)
}

func TestTruncateMap_AllValuesProcessed(t *testing.T) {
	tr := New(WithMaxLen(4))
	in := map[string]string{
		"short": "hi",
		"long":  "averylongsecret",
		"empty": "",
	}
	out := tr.TruncateMap(in)
	if out["short"] != "hi" {
		t.Errorf("short: expected %q, got %q", "hi", out["short"])
	}
	if !strings.HasSuffix(out["long"], DefaultSuffix) {
		t.Errorf("long: expected truncation suffix, got %q", out["long"])
	}
	if out["empty"] != maskFull {
		t.Errorf("empty: expected %q, got %q", maskFull, out["empty"])
	}
}

func TestTruncateMap_DoesNotMutateInput(t *testing.T) {
	tr := New(WithMaxLen(2))
	in := map[string]string{"key": "original"}
	_ = tr.TruncateMap(in)
	if in["key"] != "original" {
		t.Fatal("input map was mutated")
	}
}
