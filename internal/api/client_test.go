package api

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)
	
	// Should allow first two requests immediately
	if !rl.tryAcquire() {
		t.Error("First request should be allowed")
	}
	if !rl.tryAcquire() {
		t.Error("Second request should be allowed")
	}
	
	// Third request should be blocked
	if rl.tryAcquire() {
		t.Error("Third request should be blocked")
	}
	
	// After waiting, should allow more requests
	time.Sleep(600 * time.Millisecond)
	if !rl.tryAcquire() {
		t.Error("Request should be allowed after refill")
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{3, 3, 3},
		{-1, 0, -1},
	}
	
	for _, test := range tests {
		result := min(test.a, test.b)
		if result != test.expected {
			t.Errorf("min(%d, %d) = %d; want %d", test.a, test.b, result, test.expected)
		}
	}
}