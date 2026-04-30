// Package logger provides a thin structured logging abstraction used
// throughout vaultpipe.
//
// It wraps the standard library's [log/slog] package and exposes a small
// helper API that keeps call-sites concise:
//
//	log := logger.New(os.Stderr, logger.LevelInfo)
//	log.WithField("path", secretPath).Info("fetching secret")
//	log.WithError(err).Error("vault request failed")
//
// Supported levels (from least to most verbose): error, warn, info, debug.
// Any unrecognised level string falls back to info.
package logger
