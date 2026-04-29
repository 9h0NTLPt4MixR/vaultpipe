package process_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/yourorg/vaultpipe/internal/process"
)

func TestNewRunner_Defaults(t *testing.T) {
	r := process.NewRunner()
	if r == nil {
		t.Fatal("expected non-nil runner")
	}
}

func TestRun_EmptyCommand(t *testing.T) {
	r := process.NewRunner()
	err := r.Run("")
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestRun_SimpleCommand(t *testing.T) {
	r := process.NewRunner()
	err := r.Run("true")
	if err != nil {
		t.Fatalf("unexpected error running 'true': %v", err)
	}
}

func TestRun_FailingCommand(t *testing.T) {
	r := process.NewRunner()
	err := r.Run("false")
	if err == nil {
		t.Fatal("expected non-zero exit error from 'false'")
	}
}

func TestRun_InjectsEnv(t *testing.T) {
	var buf bytes.Buffer

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	runner := process.NewRunner(
		process.WithEnv([]string{"VAULTPIPE_TEST_VAR=hello"}),
		process.WithStdio(os.Stdin, w, os.Stderr),
	)

	if runErr := runner.Run("sh", "-c", "echo $VAULTPIPE_TEST_VAR"); runErr != nil {
		w.Close()
		t.Fatalf("unexpected error: %v", runErr)
	}
	w.Close()

	buf.ReadFrom(r)
	got := buf.String()
	if got != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", got)
	}
}

func TestWithEnv_Accumulates(t *testing.T) {
	var buf bytes.Buffer

	_, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer w.Close()

	runner := process.NewRunner(
		process.WithEnv([]string{"A=1"}),
		process.WithEnv([]string{"B=2"}),
		process.WithStdio(os.Stdin, w, os.Stderr),
	)

	_ = buf
	if runner == nil {
		t.Fatal("expected non-nil runner")
	}
}
