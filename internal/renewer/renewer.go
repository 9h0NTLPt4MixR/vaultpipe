// Package renewer provides automatic lease renewal for Vault secrets.
// It monitors secret TTLs and renews leases before they expire, ensuring
// that injected secrets remain valid for long-running processes.
package renewer

import (
	"context"
	"sync"
	"time"

	"github.com/your-org/vaultpipe/internal/logger"
)

// LeaseRenewer defines the interface for renewing a Vault lease.
type LeaseRenewer interface {
	RenewLease(ctx context.Context, leaseID string, increment int) (time.Duration, error)
}

// Renewer manages automatic renewal of Vault leases.
type Renewer struct {
	client    LeaseRenewer
	log       *logger.Logger
	graceFrac float64
	mu        sync.Mutex
	leases    map[string]context.CancelFunc
}

// Option configures a Renewer.
type Option func(*Renewer)

// WithGraceFraction sets the fraction of the TTL remaining at which renewal
// is triggered. Defaults to 0.1 (renew when 10% of TTL remains).
func WithGraceFraction(f float64) Option {
	return func(r *Renewer) {
		if f > 0 && f < 1 {
			r.graceFrac = f
		}
	}
}

// New creates a new Renewer using the provided LeaseRenewer client.
func New(client LeaseRenewer, log *logger.Logger, opts ...Option) *Renewer {
	r := &Renewer{
		client:    client,
		log:       log,
		graceFrac: 0.1,
		leases:    make(map[string]context.CancelFunc),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Track begins background renewal for the given leaseID with the supplied TTL.
// Calling Track again with the same leaseID replaces the previous watcher.
func (r *Renewer) Track(ctx context.Context, leaseID string, ttl time.Duration) {
	r.mu.Lock()
	if cancel, ok := r.leases[leaseID]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	r.leases[leaseID] = cancel
	r.mu.Unlock()

	go r.watch(child, leaseID, ttl)
}

// Stop cancels renewal for the given leaseID.
func (r *Renewer) Stop(leaseID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cancel, ok := r.leases[leaseID]; ok {
		cancel()
		delete(r.leases, leaseID)
	}
}

// StopAll cancels all active renewals.
func (r *Renewer) StopAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, cancel := range r.leases {
		cancel()
		delete(r.leases, id)
	}
}

func (r *Renewer) watch(ctx context.Context, leaseID string, ttl time.Duration) {
	delay := time.Duration(float64(ttl) * (1 - r.graceFrac))
	timer := time.NewTimer(delay)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			newTTL, err := r.client.RenewLease(ctx, leaseID, int(ttl.Seconds()))
			if err != nil {
				r.log.WithError(err).WithField("lease_id", leaseID).Error("failed to renew lease")
				return
			}
			r.log.WithField("lease_id", leaseID).WithField("new_ttl", newTTL).Info("lease renewed")
			delay = time.Duration(float64(newTTL) * (1 - r.graceFrac))
			timer.Reset(delay)
		}
	}
}
