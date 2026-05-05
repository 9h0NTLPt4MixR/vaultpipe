// Package watcher provides a polling-based secret watcher for HashiCorp Vault.
//
// It periodically fetches secrets from one or more Vault paths and invokes a
// user-supplied ChangeHandler whenever the returned values differ from the
// previously observed state. This allows long-running processes managed by
// vaultpipe to react to secret rotation without a restart.
//
// Basic usage:
//
//	w := watcher.New(
//		vaultClient,
//		func(path string, secrets map[string]string) {
//			injector.Inject(secrets)
//		},
//		log,
//		watcher.WithPaths("secret/data/myapp"),
//		watcher.WithInterval(60*time.Second),
//	)
//	go w.Run(ctx)
package watcher
