// Package snapshot provides deterministic, comparable captures of secret
// key/value maps fetched from HashiCorp Vault.
//
// A Snapshot is an immutable, content-addressed representation of a secret
// map at a particular point in time. Its SHA-256 digest can be used to
// cheaply detect whether a set of secrets has changed between two polling
// cycles without exposing the raw secret values in log output.
//
// Typical usage:
//
//	prev, _ := snapshot.New(oldSecrets)
//	curr, _ := snapshot.New(newSecrets)
//
//	if !prev.Equal(curr) {
//		changedKeys := prev.Diff(curr)
//		// re-inject or signal the child process
//	}
package snapshot
