// Package snapshot provides functionality for capturing and comparing
// secret snapshots to detect changes between polling intervals.
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// Snapshot represents an immutable point-in-time capture of a secret map.
type Snapshot struct {
	data map[string]string
	digest string
}

// New creates a new Snapshot from the provided key/value secret map.
// The digest is computed deterministically over sorted keys.
func New(secrets map[string]string) (*Snapshot, error) {
	if secrets == nil {
		secrets = map[string]string{}
	}

	copy := make(map[string]string, len(secrets))
	for k, v := range secrets {
		copy[k] = v
	}

	digest, err := computeDigest(copy)
	if err != nil {
		return nil, fmt.Errorf("snapshot: compute digest: %w", err)
	}

	return &Snapshot{data: copy, digest: digest}, nil
}

// Digest returns the hex-encoded SHA-256 digest of the snapshot contents.
func (s *Snapshot) Digest() string {
	return s.digest
}

// Data returns a shallow copy of the underlying secret map.
func (s *Snapshot) Data() map[string]string {
	out := make(map[string]string, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out
}

// Equal reports whether two snapshots have identical contents.
func (s *Snapshot) Equal(other *Snapshot) bool {
	if other == nil {
		return false
	}
	return s.digest == other.digest
}

// Diff returns the keys whose values differ between s and other.
// Keys present in one snapshot but not the other are included.
func (s *Snapshot) Diff(other *Snapshot) []string {
	seen := make(map[string]struct{})
	var changed []string

	for k, v := range s.data {
		seen[k] = struct{}{}
		if ov, ok := other.data[k]; !ok || ov != v {
			changed = append(changed, k)
		}
	}
	for k := range other.data {
		if _, ok := seen[k]; !ok {
			changed = append(changed, k)
		}
	}
	sort.Strings(changed)
	return changed
}

func computeDigest(m map[string]string) (string, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ordered := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		ordered = append(ordered, k, m[k])
	}

	b, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
