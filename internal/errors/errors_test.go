package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError(t *testing.T) {
	t.Run("NewAPIError creates error", func(t *testing.T) {
		err := NewAPIError(404, "Not Found", "Resource not found")
		assert.NotNil(t, err)
		assert.Equal(t, 404, err.StatusCode)
		assert.Equal(t, "Not Found", err.Message)
		assert.Equal(t, "Resource not found", err.Details)
	})
	
	t.Run("APIError formats message", func(t *testing.T) {
		err := NewAPIError(500, "Internal Server Error", "Database connection failed")
		assert.Contains(t, err.Error(), "500")
		assert.Contains(t, err.Error(), "Internal Server Error")
		assert.Contains(t, err.Error(), "Database connection failed")
	})
}

func TestUserError(t *testing.T) {
	t.Run("NewUserError creates error", func(t *testing.T) {
		err := NewUserError("Invalid token", "Try logging in again", ErrInvalidToken)
		assert.NotNil(t, err)
		assert.Equal(t, "Invalid token", err.Message)
		assert.Equal(t, "Try logging in again", err.Suggestion)
		assert.Equal(t, ErrInvalidToken, err.Err)
	})
	
	t.Run("UserError formats with suggestion", func(t *testing.T) {
		err := NewUserError("Authentication failed", "Run 'cu auth login'", ErrNotAuthenticated)
		assert.Contains(t, err.Error(), "Authentication failed")
		assert.Contains(t, err.Error(), "Suggestion: Run 'cu auth login'")
	})
	
	t.Run("UserError without suggestion", func(t *testing.T) {
		err := NewUserError("Something went wrong", "", nil)
		assert.Equal(t, "Something went wrong", err.Error())
	})
}

func TestHandleHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		expectedMsg string
	}{
		{
			name:        "401 Unauthorized",
			statusCode:  401,
			body:        "unauthorized",
			expectedMsg: "Authentication failed",
		},
		{
			name:        "403 Forbidden",
			statusCode:  403,
			body:        "forbidden",
			expectedMsg: "Access denied",
		},
		{
			name:        "404 Not Found",
			statusCode:  404,
			body:        "not found",
			expectedMsg: "Resource not found",
		},
		{
			name:        "429 Rate Limited",
			statusCode:  429,
			body:        "rate limited",
			expectedMsg: "Rate limit exceeded",
		},
		{
			name:        "500 Server Error",
			statusCode:  500,
			body:        "internal error",
			expectedMsg: "ClickUp service error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HandleHTTPError(tt.statusCode, tt.body)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedMsg)
		})
	}
	
	t.Run("200 OK returns nil", func(t *testing.T) {
		err := HandleHTTPError(200, "success")
		assert.NoError(t, err)
	})
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Network error",
			err:      ErrNetworkError,
			expected: true,
		},
		{
			name:     "Rate limited error",
			err:      ErrRateLimited,
			expected: true,
		},
		{
			name:     "429 API error",
			err:      NewAPIError(429, "Too Many Requests"),
			expected: true,
		},
		{
			name:     "503 API error",
			err:      NewAPIError(503, "Service Unavailable"),
			expected: true,
		},
		{
			name:     "502 API error",
			err:      NewAPIError(502, "Bad Gateway"),
			expected: true,
		},
		{
			name:     "404 API error",
			err:      NewAPIError(404, "Not Found"),
			expected: false,
		},
		{
			name:     "Regular error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("Predefined errors exist", func(t *testing.T) {
		assert.Error(t, ErrNotAuthenticated)
		assert.Error(t, ErrTokenExpired)
		assert.Error(t, ErrInvalidToken)
		assert.Error(t, ErrNetworkError)
		assert.Error(t, ErrRateLimited)
		assert.Error(t, ErrNotFound)
		assert.Error(t, ErrInvalidInput)
		assert.Error(t, ErrConfigNotFound)
	})
}