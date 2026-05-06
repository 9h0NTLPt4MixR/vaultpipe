package redactor_test

import (
	"testing"

	"github.com/youorg/vaultpipe/internal/redactor"
)

func TestNew_NoSecrets_ReturnsUnchanged(t *testing.T) {
	r := redactor.New()
	got := r.Redact("hello world")
	if got != "hello world" {
		t.Fatalf("expected unchanged string, got %q", got)
	}
}

func TestRedact_MasksRegisteredSecret(t *testing.T) {
	r := redactor.New(redactor.WithSecrets([]string{"s3cr3t"}))
	got := r.Redact("password is s3cr3t ok")
	want := "password is [REDACTED] ok"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestRedact_MultipleOccurrences(t *testing.T) {
	r := redactor.New(redactor.WithSecrets([]string{"tok"})) 
	got := r.Redact("tok and tok again")
	want := "[REDACTED] and [REDACTED] again"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestAdd_RegistersNewSecret(t *testing.T) {
	r := redactor.New()
	r.Add("newtoken")
	got := r.Redact("bearer newtoken")
	want := "bearer [REDACTED]"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestAdd_IgnoresEmptyString(t *testing.T) {
	r := redactor.New()
	r.Add("", "real")
	got := r.Redact("value=real empty=")
	want := "value=[REDACTED] empty="
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestRedactMap_MasksValues(t *testing.T) {
	r := redactor.New(redactor.WithSecrets([]string{"hunter2"}))
	input := map[string]string{
		"PASSWORD": "hunter2",
		"USER":     "alice",
	}
	out := r.RedactMap(input)
	if out["PASSWORD"] != "[REDACTED]" {
		t.Fatalf("expected PASSWORD to be redacted, got %q", out["PASSWORD"])
	}
	if out["USER"] != "alice" {
		t.Fatalf("expected USER unchanged, got %q", out["USER"])
	}
}

func TestRedactMap_OriginalUnmodified(t *testing.T) {
	r := redactor.New(redactor.WithSecrets([]string{"secret"}))
	input := map[string]string{"KEY": "secret"}
	_ = r.RedactMap(input)
	if input["KEY"] != "secret" {
		t.Fatal("original map should not be modified")
	}
}
