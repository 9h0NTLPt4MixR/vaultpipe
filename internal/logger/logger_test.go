package logger_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/yourusername/vaultpipe/internal/logger"
)

func TestNew_DefaultsToStderr(t *testing.T) {
	// Should not panic when w is nil.
	l := logger.New(nil, logger.LevelInfo)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_LevelsFiltered(t *testing.T) {
	tests := []struct {
		level    logger.Level
		msg      string
		expected bool // should the message appear?
	}{
		{logger.LevelInfo, "visible", true},
		{logger.LevelWarn, "also visible", true},
		// debug message should NOT appear when level is Info
		{logger.LevelInfo, "", false},
	}

	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo)

	l.Info("visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Error("expected info message to appear")
	}

	buf.Reset()
	l.Debug("hidden debug")
	if strings.Contains(buf.String(), "hidden debug") {
		t.Error("expected debug message to be filtered out at Info level")
	}
	_ = tests
}

func TestNew_UnrecognisedLevelDefaultsToInfo(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "verbose") // unknown level
	l.Info("hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Error("expected info message to appear with unknown level")
	}
}

func TestWithField(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelDebug)
	l2 := l.WithField("component", "injector")
	l2.Info("test field")
	if !strings.Contains(buf.String(), "component") {
		t.Error("expected field key 'component' in output")
	}
	if !strings.Contains(buf.String(), "injector") {
		t.Error("expected field value 'injector' in output")
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelDebug)
	l2 := l.WithError(errors.New("vault unreachable"))
	l2.Error("failed")
	if !strings.Contains(buf.String(), "vault unreachable") {
		t.Error("expected error message in output")
	}
}

func TestWithError_Nil(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelDebug)
	l2 := l.WithError(nil)
	if l2 == nil {
		t.Fatal("expected non-nil logger when error is nil")
	}
}
