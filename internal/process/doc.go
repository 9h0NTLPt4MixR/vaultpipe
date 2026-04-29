// Package process manages child-process lifecycle for vaultpipe.
//
// It provides a Runner that spawns a subprocess with an environment
// merged from the host and secrets injected from Vault, as well as a
// SignalForwarder that transparently relays OS signals (SIGINT, SIGTERM,
// SIGHUP) to the child so that graceful-shutdown semantics are preserved.
//
// Typical usage:
//
//	runner := process.NewRunner(
//	    process.WithEnv(secrets),  // []string{"KEY=value", ...}
//	)
//	if err := runner.Run("myapp", "--config", "app.yaml"); err != nil {
//	    log.Fatal(err)
//	}
package process
