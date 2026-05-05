// Package health provides a simple HTTP health-check handler that reports
// the liveness of vaultpipe and its connection to HashiCorp Vault.
package health

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Status represents the health state of the service.
type Status struct {
	OK        bool      `json:"ok"`
	VaultReady bool     `json:"vault_ready"`
	CheckedAt time.Time `json:"checked_at"`
	Message   string    `json:"message,omitempty"`
}

// Checker holds the current health state and exposes an HTTP handler.
type Checker struct {
	vaultReady atomic.Bool
}

// New returns a new Checker with vault_ready defaulting to false.
func New() *Checker {
	return &Checker{}
}

// SetVaultReady updates the vault connectivity flag.
func (c *Checker) SetVaultReady(ready bool) {
	c.vaultReady.Store(ready)
}

// Handler returns an http.HandlerFunc that serves the health status as JSON.
// It responds with 200 when healthy and 503 when vault is not ready.
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		vaultOK := c.vaultReady.Load()
		s := Status{
			OK:        vaultOK,
			VaultReady: vaultOK,
			CheckedAt: time.Now().UTC(),
		}
		if !vaultOK {
			s.Message = "vault connection not established"
		}

		code := http.StatusOK
		if !vaultOK {
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(s)
	}
}
