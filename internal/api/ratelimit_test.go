package api

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClock is a manually advanced clock for deterministic refill tests.
type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// newFakeClockLimiter returns a limiter driven by a fake clock that only
// moves when the test calls Advance.
func newFakeClockLimiter(maxRequests int, per time.Duration) (*RateLimiter, *fakeClock) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	rl := NewRateLimiter(maxRequests, per)
	rl.now = clk.Now
	rl.lastRefill = clk.Now()
	return rl, clk
}

// guardCtx bounds Wait calls that should return immediately, so a
// regression fails the test instead of hanging it.
func guardCtx(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestNewRateLimiter(t *testing.T) {
	t.Run("creates rate limiter with correct parameters", func(t *testing.T) {
		rl := NewRateLimiter(10, time.Second)
		assert.NotNil(t, rl)
		assert.Equal(t, 10, rl.tokens)
		assert.Equal(t, 10, rl.maxTokens)
		assert.Equal(t, 100*time.Millisecond, rl.refillRate)
	})

	t.Run("different rates", func(t *testing.T) {
		tests := []struct {
			maxRequests int
			per         time.Duration
			wantRefill  time.Duration
		}{
			{100, time.Minute, 600 * time.Millisecond},
			{60, time.Minute, time.Second},
			{1, time.Second, time.Second},
			{10, 100 * time.Millisecond, 10 * time.Millisecond},
		}

		for _, tt := range tests {
			rl := NewRateLimiter(tt.maxRequests, tt.per)
			assert.Equal(t, tt.wantRefill, rl.refillRate)
		}
	})
}

func TestRateLimiterWait(t *testing.T) {
	t.Run("allows burst up to limit", func(t *testing.T) {
		// Clock is frozen, so these can only succeed from the initial burst
		rl, _ := newFakeClockLimiter(3, time.Second)
		ctx := guardCtx(t)

		for i := 0; i < 3; i++ {
			assert.NoError(t, rl.Wait(ctx))
		}
		assert.False(t, rl.tryAcquire(), "Burst should be exhausted")
	})

	t.Run("waits when limit exceeded", func(t *testing.T) {
		rl, clk := newFakeClockLimiter(2, 20*time.Millisecond)
		ctx := guardCtx(t)

		// Use up tokens
		require.NoError(t, rl.Wait(ctx))
		require.NoError(t, rl.Wait(ctx))

		done := make(chan error, 1)
		go func() { done <- rl.Wait(ctx) }()

		// Clock is frozen, so Wait cannot return yet
		select {
		case err := <-done:
			t.Fatalf("Wait returned before refill: %v", err)
		case <-time.After(50 * time.Millisecond):
		}

		// One refill period unblocks it on the next poll
		clk.Advance(10 * time.Millisecond)
		select {
		case err := <-done:
			assert.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("Wait did not return after refill")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		rl := NewRateLimiter(1, time.Hour) // Very slow refill
		ctx, cancel := context.WithCancel(context.Background())

		// Use up the token
		require.NoError(t, rl.Wait(ctx))

		// Cancel context while waiting
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		err := rl.Wait(ctx)
		elapsed := time.Since(start)

		assert.ErrorIs(t, err, context.Canceled)
		// Generous bound: refill is an hour, so returning at all proves cancellation
		assert.Less(t, elapsed, 5*time.Second)
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		rl, clk := newFakeClockLimiter(2, 100*time.Millisecond)
		ctx := guardCtx(t)

		// Use all tokens
		require.NoError(t, rl.Wait(ctx))
		require.NoError(t, rl.Wait(ctx))

		// One refill period; with the clock frozen otherwise, Wait can only
		// succeed from the refilled token
		clk.Advance(50 * time.Millisecond)

		assert.NoError(t, rl.Wait(ctx))
		assert.False(t, rl.tryAcquire(), "Only one token should have refilled")
	})
}

func TestRateLimiterConcurrency(t *testing.T) {
	t.Run("handles concurrent requests safely", func(t *testing.T) {
		// Frozen clock: no refills, so exactly the burst succeeds and the
		// rest block until the context expires
		rl, _ := newFakeClockLimiter(10, 100*time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		var wg sync.WaitGroup
		var successCount int32
		numGoroutines := 20

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := rl.Wait(ctx); err == nil {
					atomic.AddInt32(&successCount, 1)
				}
			}()
		}

		wg.Wait()

		assert.Equal(t, int32(10), atomic.LoadInt32(&successCount))
	})

	t.Run("maintains rate limit under load", func(t *testing.T) {
		// Real time; only a lower bound is asserted, which sleep overshoot
		// cannot violate
		rl := NewRateLimiter(5, 200*time.Millisecond)
		ctx := context.Background()

		start := time.Now()
		requestCount := 0

		// Try to make 10 requests
		for i := 0; i < 10; i++ {
			if err := rl.Wait(ctx); err == nil {
				requestCount++
			}
		}

		elapsed := time.Since(start)

		// Should have made all 10 requests
		assert.Equal(t, 10, requestCount)

		// Should have taken at least 200ms to complete
		// (5 immediate, then wait ~40ms, get 1, wait ~40ms, etc)
		assert.True(t, elapsed >= 200*time.Millisecond, "Should respect rate limit timing")
	})
}

func TestTryAcquire(t *testing.T) {
	t.Run("acquires tokens correctly", func(t *testing.T) {
		rl := NewRateLimiter(3, time.Second)

		// Should succeed for available tokens
		assert.True(t, rl.tryAcquire())
		assert.True(t, rl.tryAcquire())
		assert.True(t, rl.tryAcquire())

		// Should fail when no tokens
		assert.False(t, rl.tryAcquire())
	})

	t.Run("refills tokens correctly", func(t *testing.T) {
		rl, clk := newFakeClockLimiter(2, 100*time.Millisecond)

		// Use all tokens
		assert.True(t, rl.tryAcquire())
		assert.True(t, rl.tryAcquire())
		assert.False(t, rl.tryAcquire())

		// Just short of one refill period
		clk.Advance(49 * time.Millisecond)
		assert.False(t, rl.tryAcquire())

		// Complete one refill period
		clk.Advance(1 * time.Millisecond)

		// Should have 1 token
		assert.True(t, rl.tryAcquire())
		assert.False(t, rl.tryAcquire())

		// Another refill period
		clk.Advance(50 * time.Millisecond)

		// Should have 1 more token (not exceeding max)
		assert.True(t, rl.tryAcquire())
		assert.False(t, rl.tryAcquire())
	})

	t.Run("does not exceed max tokens", func(t *testing.T) {
		rl, clk := newFakeClockLimiter(2, 100*time.Millisecond)

		// Long enough for multiple refills
		clk.Advance(300 * time.Millisecond)

		// Should still be capped at max
		assert.True(t, rl.tryAcquire())
		assert.True(t, rl.tryAcquire())
		assert.False(t, rl.tryAcquire())
	})
}
