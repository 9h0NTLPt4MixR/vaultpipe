package sanitizer_test

import (
	"testing"

	"github.com/your-org/vaultpipe/internal/sanitizer"
)

func TestSanitize_ValidKey(t *testing.T) {
	s := sanitizer.New()
	got, err := s.Sanitize("my_secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "MY_SECRET" {
		t.Fatalf("want MY_SECRET, got %s", got)
	}
}

func TestSanitize_EmptyKey(t *testing.T) {
	s := sanitizer.New()
	_, err := s.Sanitize("")
	if err != sanitizer.ErrEmptyKey {
		t.Fatalf("want ErrEmptyKey, got %v", err)
	}
}

func TestSanitize_ReplacesInvalidChars(t *testing.T) {
	s := sanitizer.New(sanitizer.WithStrict(false))
	got, err := s.Sanitize("my-secret/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "MY_SECRET_PATH" {
		t.Fatalf("want MY_SECRET_PATH, got %s", got)
	}
}

func TestSanitize_StrictRejectsInvalidKey(t *testing.T) {
	s := sanitizer.New(sanitizer.WithStrict(true))
	_, err := s.Sanitize("bad-key")
	if err != sanitizer.ErrInvalidKey {
		t.Fatalf("want ErrInvalidKey, got %v", err)
	}
}

func TestSanitize_StrictAcceptsValidKey(t *testing.T) {
	s := sanitizer.New(sanitizer.WithStrict(true))
	got, err := s.Sanitize("GOOD_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "GOOD_KEY" {
		t.Fatalf("want GOOD_KEY, got %s", got)
	}
}

func TestSanitize_LeadingDigitPrefixed(t *testing.T) {
	s := sanitizer.New(sanitizer.WithUppercase(false))
	got, err := s.Sanitize("1bad")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "_1bad" {
		t.Fatalf("want _1bad, got %s", got)
	}
}

func TestSanitize_WithoutUppercase(t *testing.T) {
	s := sanitizer.New(sanitizer.WithUppercase(false))
	got, err := s.Sanitize("myKey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "myKey" {
		t.Fatalf("want myKey, got %s", got)
	}
}

func TestSanitizeMap_NormalisesKeys(t *testing.T) {
	s := sanitizer.New()
	input := map[string]string{
		"db-host": "localhost",
		"db-port": "5432",
	}
	out, err := s.SanitizeMap(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"DB_HOST", "DB_PORT"} {
		if _, ok := out[want]; !ok {
			t.Errorf("missing expected key %s", want)
		}
	}
}

func TestSanitizeMap_StrictReturnsError(t *testing.T) {
	s := sanitizer.New(sanitizer.WithStrict(true))
	input := map[string]string{"bad-key": "value"}
	_, err := s.SanitizeMap(input)
	if err == nil {
		t.Fatal("expected error for invalid key in strict mode")
	}
}
