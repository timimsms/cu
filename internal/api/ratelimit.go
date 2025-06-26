package api

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, per time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:     maxRequests,
		maxTokens:  maxRequests,
		refillRate: per / time.Duration(maxRequests),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		if r.tryAcquire() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.refillRate):
			// Try again after waiting
		}
	}
}

// tryAcquire attempts to acquire a token
func (r *RateLimiter) tryAcquire() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed / r.refillRate)

	if tokensToAdd > 0 {
		r.tokens = min(r.tokens+tokensToAdd, r.maxTokens)
		r.lastRefill = now
	}

	// Try to acquire a token
	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
