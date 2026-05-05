package metrics

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.HandlerFunc that serves a JSON snapshot of m.
// It is intended to be mounted on a lightweight debug/admin HTTP mux.
//
// Example:
//
//	http.Handle("/metrics", metrics.Handler(m))
func Handler(m *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		snap := m.Snapshot()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(snap)
	}
}
