package template_test

import (
	"testing"

	"github.com/your-org/vaultpipe/internal/template"
)

func TestRender_StaticPath(t *testing.T) {
	r := template.New()
	got, err := r.Render("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "secret/data/myapp" {
		t.Errorf("got %q, want %q", got, "secret/data/myapp")
	}
}

func TestRender_WithExtraVars(t *testing.T) {
	r := template.New(template.WithVars(map[string]string{
		"REGION": "us-east-1",
	}))
	got, err := r.Render("secret/data/{{ .Var.REGION }}/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "secret/data/us-east-1/db"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRender_WithEnvVar(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	r := template.New()
	got, err := r.Render("secret/data/{{ .Env.APP_ENV }}/config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "secret/data/production/config"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRender_MissingVarReturnsError(t *testing.T) {
	r := template.New()
	_, err := r.Render("secret/{{ .Var.UNDEFINED_KEY }}")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestRender_InvalidTemplateSyntax(t *testing.T) {
	r := template.New()
	_, err := r.Render("secret/{{ .Var.UNCLOSED")
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestWithVars_MultipleCallsMerge(t *testing.T) {
	r := template.New(
		template.WithVars(map[string]string{"A": "alpha"}),
		template.WithVars(map[string]string{"B": "beta"}),
	)
	got, err := r.Render("{{ .Var.A }}/{{ .Var.B }}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "alpha/beta" {
		t.Errorf("got %q, want %q", got, "alpha/beta")
	}
}
