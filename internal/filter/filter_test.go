package filter_test

import (
	"testing"

	"github.com/your-org/vaultpipe/internal/filter"
)

func baseSecrets() map[string]string {
	return map[string]string{
		"DB_HOST":     "localhost",
		"DB_PASSWORD": "secret",
		"API_KEY":     "abc123",
		"LOG_LEVEL":   "info",
	}
}

func TestApply_NoOptions_ReturnsAll(t *testing.T) {
	f := filter.New()
	out := f.Apply(baseSecrets())
	if len(out) != 4 {
		t.Fatalf("expected 4 keys, got %d", len(out))
	}
}

func TestApply_Allowlist_ReturnsOnlyAllowed(t *testing.T) {
	f := filter.New(filter.WithAllowlist("DB_HOST", "API_KEY"))
	out := f.Apply(baseSecrets())
	if len(out) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(out))
	}
	if _, ok := out["DB_HOST"]; !ok {
		t.Error("expected DB_HOST in output")
	}
	if _, ok := out["API_KEY"]; !ok {
		t.Error("expected API_KEY in output")
	}
}

func TestApply_Denylist_ExcludesDenied(t *testing.T) {
	f := filter.New(filter.WithDenylist("DB_PASSWORD"))
	out := f.Apply(baseSecrets())
	if _, ok := out["DB_PASSWORD"]; ok {
		t.Error("DB_PASSWORD should have been excluded")
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(out))
	}
}

func TestApply_KeyPrefix_FiltersCorrectly(t *testing.T) {
	f := filter.New(filter.WithKeyPrefix("DB_"))
	out := f.Apply(baseSecrets())
	if len(out) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(out))
	}
	for k := range out {
		if k != "DB_HOST" && k != "DB_PASSWORD" {
			t.Errorf("unexpected key %q in output", k)
		}
	}
}

func TestApply_DenylistOverridesAllowlist(t *testing.T) {
	f := filter.New(
		filter.WithAllowlist("DB_HOST", "DB_PASSWORD"),
		filter.WithDenylist("DB_PASSWORD"),
	)
	out := f.Apply(baseSecrets())
	if _, ok := out["DB_PASSWORD"]; ok {
		t.Error("DB_PASSWORD should be denied even when allowlisted")
	}
	if _, ok := out["DB_HOST"]; !ok {
		t.Error("expected DB_HOST in output")
	}
}

func TestApply_EmptyInput_ReturnsEmpty(t *testing.T) {
	f := filter.New(filter.WithAllowlist("ANY_KEY"))
	out := f.Apply(map[string]string{})
	if len(out) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(out))
	}
}

func TestApply_PrefixAndDenylist_Combined(t *testing.T) {
	f := filter.New(
		filter.WithKeyPrefix("DB_"),
		filter.WithDenylist("DB_PASSWORD"),
	)
	out := f.Apply(baseSecrets())
	if len(out) != 1 {
		t.Fatalf("expected 1 key, got %d", len(out))
	}
	if _, ok := out["DB_HOST"]; !ok {
		t.Error("expected DB_HOST in output")
	}
}
