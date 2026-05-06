// Package ratelimit implements a thread-safe token-bucket rate limiter
// designed to throttle outbound requests to HashiCorp Vault.
//
// # Overview
//
// A Limiter is created with a sustained rate (tokens per second) and an
// optional burst capacity. Callers use Wait to block until a token is
// available, or TryAcquire for a non-blocking attempt.
//
// # Usage
//
//	limiter, err := ratelimit.New(
//		ratelimit.WithRate(5),   // 5 requests per second
//		ratelimit.WithBurst(10), // allow bursts of up to 10
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if err := limiter.Wait(ctx); err != nil {
//		return err
//	}
//	// safe to call Vault now
package ratelimit
