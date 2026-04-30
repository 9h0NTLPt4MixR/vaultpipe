// Package logger provides a structured leveled logger for vaultpipe.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Level represents a logging verbosity level.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Logger wraps slog.Logger with vaultpipe-specific helpers.
type Logger struct {
	*slog.Logger
}

// New creates a Logger writing to w at the given level.
// If w is nil, os.Stderr is used. If level is unrecognised, Info is used.
func New(w io.Writer, level Level) *Logger {
	if w == nil {
		w = os.Stderr
	}

	var sl slog.Level
	switch strings.ToLower(string(level)) {
	case string(LevelDebug):
		sl = slog.LevelDebug
	case string(LevelWarn):
		sl = slog.LevelWarn
	case string(LevelError):
		sl = slog.LevelError
	default:
		sl = slog.LevelInfo
	}

	h := slog.NewTextHandler(w, &slog.HandlerOptions{Level: sl})
	return &Logger{slog.New(h)}
}

// WithField returns a new Logger with a permanent key/value pair attached.
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{l.Logger.With(key, value)}
}

// WithError returns a new Logger with the error attached as a field.
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}
