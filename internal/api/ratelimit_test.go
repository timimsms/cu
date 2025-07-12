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
		rl := NewRateLimiter(3, time.Second)
		ctx := context.Background()

		// Should allow 3 immediate requests
		for i := 0; i < 3; i++ {
			start := time.Now()
			err := rl.Wait(ctx)
			elapsed := time.Since(start)
			
			assert.NoError(t, err)
			assert.Less(t, elapsed, 10*time.Millisecond, "Should not wait for burst")
		}
	})

	t.Run("waits when limit exceeded", func(t *testing.T) {
		rl := NewRateLimiter(2, 200*time.Millisecond)
		ctx := context.Background()

		// Use up tokens
		require.NoError(t, rl.Wait(ctx))
		require.NoError(t, rl.Wait(ctx))

		// Next request should wait
		start := time.Now()
		err := rl.Wait(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
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

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Less(t, elapsed, 100*time.Millisecond)
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		rl := NewRateLimiter(2, 100*time.Millisecond)
		ctx := context.Background()

		// Use all tokens
		require.NoError(t, rl.Wait(ctx))
		require.NoError(t, rl.Wait(ctx))

		// Wait for refill
		time.Sleep(60 * time.Millisecond)

		// Should have 1 token refilled
		start := time.Now()
		err := rl.Wait(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, elapsed, 10*time.Millisecond, "Should not wait after refill")
	})
}

func TestRateLimiterConcurrency(t *testing.T) {
	t.Run("handles concurrent requests safely", func(t *testing.T) {
		rl := NewRateLimiter(10, 100*time.Millisecond)
		ctx := context.Background()

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

		// Should have allowed exactly 10 requests immediately
		// Others would need to wait for refill
		assert.GreaterOrEqual(t, atomic.LoadInt32(&successCount), int32(10))
	})

	t.Run("maintains rate limit under load", func(t *testing.T) {
		// This test is inherently timing-sensitive
		// We'll use a larger window to reduce flakiness
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
		rl := NewRateLimiter(2, 100*time.Millisecond)

		// Use all tokens
		assert.True(t, rl.tryAcquire())
		assert.True(t, rl.tryAcquire())
		assert.False(t, rl.tryAcquire())

		// Wait for one refill period
		time.Sleep(55 * time.Millisecond)
		
		// Should have 1 token
		assert.True(t, rl.tryAcquire())
		assert.False(t, rl.tryAcquire())

		// Wait for full refill
		time.Sleep(55 * time.Millisecond)
		
		// Should have 1 more token (not exceeding max)
		assert.True(t, rl.tryAcquire())
		assert.False(t, rl.tryAcquire())
	})

	t.Run("does not exceed max tokens", func(t *testing.T) {
		rl := NewRateLimiter(2, 100*time.Millisecond)

		// Wait long enough for multiple refills
		time.Sleep(300 * time.Millisecond)

		// Should still be capped at max
		assert.True(t, rl.tryAcquire())
		assert.True(t, rl.tryAcquire())
		assert.False(t, rl.tryAcquire())
	})
}