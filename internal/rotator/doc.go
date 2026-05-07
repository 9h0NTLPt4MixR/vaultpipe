// Package rotator provides coordinated secret rotation for vaultpipe.
//
// A Rotator watches a Vault lease expiry time and, when the remaining
// lifetime falls within a configurable threshold window, invokes a
// caller-supplied RotateFunc to re-fetch and re-inject secrets.
//
// Typical usage:
//
//	rot := rotator.New(log,
//		rotator.WithInterval(30*time.Second),
//		rotator.WithThreshold(5*time.Minute),
//	)
//
//	go rot.Run(ctx, leaseExpiry, func(ctx context.Context) error {
//		secrets, err := vaultClient.GetSecrets(ctx, path)
//		if err != nil {
//			return err
//		}
//		return injector.Inject(secrets)
//	})
//
// The rotation loop exits cleanly when the provided context is cancelled,
// returning the context's error to the caller.
package rotator
