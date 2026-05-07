package resolver_test

import (
	"testing"

	"github.com/your-org/vaultpipe/internal/resolver"
)

func staticEnv(m map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := m[key]
		return v, ok
	}
}

func TestResolve_StaticPath(t *testing.T) {
	r := resolver.New()
	got, err := r.Resolve("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "secret/data/myapp" {
		t.Errorf("expected unchanged path, got %q", got)
	}
}

func TestResolve_ExpandsEnvVar(t *testing.T) {
	r := resolver.New(resolver.WithEnvLookup(staticEnv(map[string]string{
		"APP_ENV": "production",
	})))

	got, err := r.Resolve("secret/data/${APP_ENV}/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "secret/data/production/myapp"; got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestResolve_MissingVarReturnsError(t *testing.T) {
	r := resolver.New(resolver.WithEnvLookup(staticEnv(map[string]string{})))

	_, err := r.Resolve("secret/data/${UNDEFINED}/myapp")
	if err == nil {
		t.Fatal("expected error for missing variable, got nil")
	}
}

func TestResolve_MultipleVars(t *testing.T) {
	r := resolver.New(resolver.WithEnvLookup(staticEnv(map[string]string{
		"ORG":  "acme",
		"TEAM": "platform",
	})))

	got, err := r.Resolve("secret/${ORG}/${TEAM}/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "secret/acme/platform/db"; got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestResolveAll_AllExpanded(t *testing.T) {
	r := resolver.New(resolver.WithEnvLookup(staticEnv(map[string]string{
		"ENV": "staging",
	})))

	paths := []string{"secret/data/${ENV}/svc-a", "secret/data/${ENV}/svc-b"}
	got, err := r.ResolveAll(paths)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(got))
	}
	if got[0] != "secret/data/staging/svc-a" || got[1] != "secret/data/staging/svc-b" {
		t.Errorf("unexpected resolved paths: %v", got)
	}
}

func TestResolveAll_StopsOnFirstError(t *testing.T) {
	r := resolver.New(resolver.WithEnvLookup(staticEnv(map[string]string{})))

	_, err := r.ResolveAll([]string{"secret/${MISSING}/a", "secret/static/b"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
