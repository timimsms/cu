package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/errors"
)

// MockAuthManager implements the auth manager interface for testing
type MockAuthManager struct {
	token   *auth.Token
	err     error
	callLog []string
}

func (m *MockAuthManager) GetCurrentToken() (*auth.Token, error) {
	m.callLog = append(m.callLog, "GetCurrentToken")
	if m.err != nil {
		return nil, m.err
	}
	return m.token, nil
}

func (m *MockAuthManager) Reset() {
	m.callLog = []string{}
}

func TestNewClient(t *testing.T) {
	t.Run("creates client with auth manager", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}

		client := NewClient(authManager)
		
		assert.NotNil(t, client)
		assert.Equal(t, authManager, client.authManager)
		assert.NotNil(t, client.rateLimiter)
		assert.Nil(t, client.client) // Not connected yet
		assert.Nil(t, client.userLookup) // Not connected yet
	})
}

func TestClient_Connect(t *testing.T) {
	t.Run("connects successfully with valid token", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)

		err := client.Connect()
		
		assert.NoError(t, err)
		assert.NotNil(t, client.client)
		assert.NotNil(t, client.userLookup)
		assert.Contains(t, authManager.callLog, "GetCurrentToken")
	})

	t.Run("fails when auth manager returns error", func(t *testing.T) {
		authManager := &MockAuthManager{
			err: fmt.Errorf("auth error"),
		}
		client := NewClient(authManager)

		err := client.Connect()
		
		assert.Error(t, err)
		assert.Equal(t, errors.ErrNotAuthenticated, err)
		assert.Nil(t, client.client)
	})

	t.Run("fails when no token available", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: nil,
			err:   fmt.Errorf("no token"),
		}
		client := NewClient(authManager)

		err := client.Connect()
		
		assert.Error(t, err)
		assert.Equal(t, errors.ErrNotAuthenticated, err)
	})
}

func TestClient_UserLookup(t *testing.T) {
	t.Run("returns user lookup service", func(t *testing.T) {
		client := &Client{
			userLookup: &UserLookup{},
		}

		lookup := client.UserLookup()
		
		assert.NotNil(t, lookup)
		assert.Equal(t, client.userLookup, lookup)
	})

	t.Run("returns nil when not initialized", func(t *testing.T) {
		client := &Client{}

		lookup := client.UserLookup()
		
		assert.Nil(t, lookup)
	})
}

// Test error handling
func TestClient_HandleError(t *testing.T) {
	client := &Client{}

	t.Run("returns nil for nil error", func(t *testing.T) {
		err := client.handleError(nil)
		assert.NoError(t, err)
	})

	t.Run("returns original error", func(t *testing.T) {
		originalErr := fmt.Errorf("test error")
		err := client.handleError(originalErr)
		assert.Equal(t, originalErr, err)
	})
}

// Test parseDueDate function (internal function in package)
func TestParseDueDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "ISO date",
			input:    "2022-01-01",
			hasError: false,
		},
		{
			name:     "RFC3339 date",
			input:    "2022-01-01T15:04:05Z",
			hasError: false,
		},
		{
			name:     "today keyword",
			input:    "today",
			hasError: false,
		},
		{
			name:     "tomorrow keyword",
			input:    "tomorrow",
			hasError: false,
		},
		{
			name:     "week keyword",
			input:    "week",
			hasError: false,
		},
		{
			name:     "Invalid format",
			input:    "invalid-date",
			hasError: true,
		},
		{
			name:     "Empty string",
			input:    "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseDueDate(tt.input)
			
			if tt.hasError {
				assert.Error(t, err)
				assert.True(t, result.IsZero())
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}

// Test rate limiter functionality
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

// Test rate limiter with context cancellation
func TestRateLimiterWithContext(t *testing.T) {
	rl := NewRateLimiter(1, time.Second)
	ctx := context.Background()

	t.Run("allows request when not rate limited", func(t *testing.T) {
		err := rl.Wait(ctx)
		assert.NoError(t, err)
	})

	t.Run("waits when rate limited", func(t *testing.T) {
		// Fill up the bucket
		rl.tryAcquire()
		
		// This should wait but not error
		start := time.Now()
		err := rl.Wait(ctx)
		elapsed := time.Since(start)
		
		assert.NoError(t, err)
		assert.True(t, elapsed >= 100*time.Millisecond, "Should have waited")
	})

	t.Run("returns error when context is cancelled", func(t *testing.T) {
		// Fill up the bucket
		rl.tryAcquire()
		
		// Cancel context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		
		err := rl.Wait(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

// Test client initialization with rate limiter
func TestClientRateLimiter(t *testing.T) {
	t.Run("client has functioning rate limiter", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		
		// Test that rate limiter is working
		assert.NotNil(t, client.rateLimiter)
		
		// Test rate limiter functionality
		ctx := context.Background()
		err := client.rateLimiter.Wait(ctx)
		assert.NoError(t, err)
	})
}

// Test client authentication flow
func TestClientAuthFlow(t *testing.T) {
	t.Run("multiple connect calls reuse auth", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)

		// First connect
		err1 := client.Connect()
		assert.NoError(t, err1)
		
		// Second connect (should work fine)
		err2 := client.Connect()
		assert.NoError(t, err2)
		
		// Should have called GetCurrentToken twice
		assert.Equal(t, 2, len(authManager.callLog))
		assert.Equal(t, "GetCurrentToken", authManager.callLog[0])
		assert.Equal(t, "GetCurrentToken", authManager.callLog[1])
	})

	t.Run("connect with nil token fails", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: nil, // nil token
			err:   fmt.Errorf("no token available"),
		}
		client := NewClient(authManager)

		err := client.Connect()
		
		assert.Error(t, err)
		assert.Equal(t, errors.ErrNotAuthenticated, err)
	})
}

// Test client structure and initialization
func TestClientStructure(t *testing.T) {
	t.Run("new client has expected structure", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		
		// Check initial state
		assert.NotNil(t, client.authManager)
		assert.NotNil(t, client.rateLimiter)
		assert.Nil(t, client.client)
		assert.Nil(t, client.userLookup)
		
		// After connect
		err := client.Connect()
		assert.NoError(t, err)
		
		assert.NotNil(t, client.client)
		assert.NotNil(t, client.userLookup)
	})
}