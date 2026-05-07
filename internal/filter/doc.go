// Package filter provides composable key-based filtering for secrets maps
// retrieved from HashiCorp Vault.
//
// A Filter can be configured with an allowlist (only these keys pass),
// a denylist (these keys are always excluded), and a key prefix (only keys
// starting with the prefix are retained). Rules are evaluated in the
// following order:
//
//  1. Denylist — matching keys are unconditionally removed.
//  2. Prefix   — keys not matching the prefix are removed.
//  3. Allowlist — when non-empty, only listed keys are retained.
//
// Filters are safe for concurrent reads once constructed; they must not
// be mutated after the first call to Apply.
package filter
