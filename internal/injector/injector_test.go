package injector

import (
	"os"
	"testing"
)

func TestNewInjector_Defaults(t *testing.T) {
	inj := NewInjector()
	if inj.prefix != "" {
		t.Errorf("expected empty prefix, got %q", inj.prefix)
	}
	if inj.overwrite {
		t.Error("expected overwrite to be false by default")
	}
}

func TestNewInjector_WithOptions(t *testing.T) {
	inj := NewInjector(WithPrefix("myapp"), WithOverwrite(true))
	if inj.prefix != "MYAPP" {
		t.Errorf("expected prefix MYAPP, got %q", inj.prefix)
	}
	if !inj.overwrite {
		t.Error("expected overwrite to be true")
	}
}

func TestInject_SetsEnvVars(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("API_KEY")
	})

	inj := NewInjector()
	secrets := map[string]string{
		"db_password": "supersecret",
		"api_key":     "abc123",
	}

	if err := inj.Inject(secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("DB_PASSWORD"); got != "supersecret" {
		t.Errorf("expected DB_PASSWORD=supersecret, got %q", got)
	}
	if got := os.Getenv("API_KEY"); got != "abc123" {
		t.Errorf("expected API_KEY=abc123, got %q", got)
	}
}

func TestInject_WithPrefix(t *testing.T) {
	t.Cleanup(func() { os.Unsetenv("VAULT_TOKEN") })

	inj := NewInjector(WithPrefix("vault"))
	if err := inj.Inject(map[string]string{"token": "mytoken"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv("VAULT_TOKEN"); got != "mytoken" {
		t.Errorf("expected VAULT_TOKEN=mytoken, got %q", got)
	}
}

func TestInject_NoOverwrite(t *testing.T) {
	os.Setenv("EXISTING_KEY", "original")
	t.Cleanup(func() { os.Unsetenv("EXISTING_KEY") })

	inj := NewInjector() // overwrite=false by default
	if err := inj.Inject(map[string]string{"existing_key": "new_value"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv("EXISTING_KEY"); got != "original" {
		t.Errorf("expected EXISTING_KEY to remain 'original', got %q", got)
	}
}

func TestInjectToMap_DoesNotMutateEnv(t *testing.T) {
	inj := NewInjector(WithPrefix("app"))
	result := inj.InjectToMap(map[string]string{"secret": "val"})

	if got, ok := result["APP_SECRET"]; !ok || got != "val" {
		t.Errorf("expected APP_SECRET=val in result map, got %q", got)
	}
	if os.Getenv("APP_SECRET") != "" {
		t.Error("InjectToMap should not modify the process environment")
	}
}
