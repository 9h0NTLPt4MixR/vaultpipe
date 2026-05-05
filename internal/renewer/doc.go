// Package renewer provides automatic Vault lease renewal for long-running
// processes managed by vaultpipe.
//
// When Vault issues a secret with a finite TTL, the renewer package tracks
// the associated lease and schedules a background goroutine to call the Vault
// renew endpoint before the lease expires. This prevents credential
// invalidation without requiring a full re-fetch of secrets.
//
// Usage:
//
//	client := vault.NewClient(cfg)
//	r := renewer.New(client, log, renewer.WithGraceFraction(0.15))
//	r.Track(ctx, leaseID, ttl)
//	defer r.StopAll()
//
// The grace fraction controls how early renewal is attempted relative to the
// remaining TTL. A value of 0.15 means renewal fires when 15% of the TTL is
// left. The default is 0.10.
package renewer
