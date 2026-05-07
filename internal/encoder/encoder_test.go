package encoder_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/your-org/vaultpipe/internal/encoder"
)

func TestNew_DefaultsToRaw(t *testing.T) {
	e := encoder.New()
	got, err := e.Encode("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

func TestEncode_Base64Std(t *testing.T) {
	e := encoder.New(encoder.WithEncoding(encoder.Base64Std))
	const input = "super-secret"
	want := base64.StdEncoding.EncodeToString([]byte(input))
	got, err := e.Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEncode_Base64URL(t *testing.T) {
	e := encoder.New(encoder.WithEncoding(encoder.Base64URL))
	const input = "url+safe/value=="
	want := base64.URLEncoding.EncodeToString([]byte(input))
	got, err := e.Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEncode_Hex(t *testing.T) {
	e := encoder.New(encoder.WithEncoding(encoder.Hex))
	const input = "binary\x00data"
	want := hex.EncodeToString([]byte(input))
	got, err := e.Encode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEncodeMap_TransformsAllValues(t *testing.T) {
	e := encoder.New(encoder.WithEncoding(encoder.Base64Std))
	input := map[string]string{
		"DB_PASS": "s3cr3t",
		"API_KEY": "key-abc",
	}
	out, err := e.EncodeMap(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for k, v := range input {
		want := base64.StdEncoding.EncodeToString([]byte(v))
		if out[k] != want {
			t.Errorf("key %q: expected %q, got %q", k, want, out[k])
		}
	}
}

func TestEncodeMap_DoesNotMutateInput(t *testing.T) {
	e := encoder.New(encoder.WithEncoding(encoder.Hex))
	input := map[string]string{"KEY": "value"}
	_, err := e.EncodeMap(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input["KEY"] != "value" {
		t.Error("EncodeMap must not mutate the input map")
	}
}

func TestEncodeMap_EmptyMap(t *testing.T) {
	e := encoder.New(encoder.WithEncoding(encoder.Base64Std))
	out, err := e.EncodeMap(map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty map, got %v", out)
	}
}
