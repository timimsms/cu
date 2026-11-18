package errors

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAPIError(t *testing.T) {
	cause := errors.New("sdk error")

	tests := []struct {
		name       string
		statusCode int
		ecode      string
		message    string
		resource   string
		resourceID string
		kind       error
		want       string
	}{
		{
			name:       "401 token invalid",
			statusCode: 401, ecode: "OAUTH_025", message: "Token invalid",
			kind: ErrNotAuthenticated,
			want: "not authenticated or token invalid — run `cu auth login` (HTTP 401 OAUTH_025: Token invalid)",
		},
		{
			name:       "401 team not authorized is not a login problem",
			statusCode: 401, ecode: "OAUTH_027", message: "Team not authorized",
			resource: "task", resourceID: "bogus",
			kind: ErrNotFound,
			want: `task "bogus" not found, or not accessible with the current token (HTTP 401 OAUTH_027: Team not authorized)`,
		},
		{
			name:       "403 with resource",
			statusCode: 403, resource: "list", resourceID: "42",
			kind: ErrPermissionDenied,
			want: `permission denied for list "42" — check your ClickUp access to it (HTTP 403)`,
		},
		{
			name:       "404 with resource",
			statusCode: 404, ecode: "ITEM_013", message: "Task not found",
			resource: "task", resourceID: "abc",
			kind: ErrNotFound,
			want: `task "abc" not found (HTTP 404 ITEM_013: Task not found)`,
		},
		{
			name:       "404 without resource",
			statusCode: 404,
			kind:       ErrNotFound,
			want:       "resource not found (HTTP 404)",
		},
		{
			name:       "429",
			statusCode: 429,
			kind:       ErrRateLimited,
			want:       "ClickUp rate limit exceeded — wait a minute and try again (HTTP 429)",
		},
		{
			name:       "500",
			statusCode: 500, message: "Internal error",
			kind: ErrServer,
			want: "ClickUp server error — try again shortly (HTTP 500: Internal error)",
		},
		{
			name:       "400",
			statusCode: 400, ecode: "INPUT_005", message: "Status not found",
			kind: ErrInvalidInput,
			want: "ClickUp rejected the request (HTTP 400 INPUT_005: Status not found)",
		},
		{
			name:       "other 4xx",
			statusCode: 409,
			want:       "ClickUp request failed (HTTP 409)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.statusCode, tt.ecode, tt.message, cause)
			err.Resource, err.ResourceID = tt.resource, tt.resourceID

			assert.Equal(t, tt.want, err.Error())
			assert.ErrorIs(t, err, cause, "original error must stay reachable")
			if tt.kind != nil {
				assert.ErrorIs(t, err, tt.kind)
			}
		})
	}
}

func TestAPIErrorRetryAfter(t *testing.T) {
	err := NewAPIError(429, "", "", nil)
	err.RetryAfter = 42*time.Second + 300*time.Millisecond
	assert.Equal(t, "ClickUp rate limit exceeded — wait 42s and try again (HTTP 429)", err.Error())
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
			err:      NewAPIError(429, "", "Too Many Requests", nil),
			expected: true,
		},
		{
			name:     "503 API error",
			err:      NewAPIError(503, "", "Service Unavailable", nil),
			expected: true,
		},
		{
			name:     "502 API error",
			err:      NewAPIError(502, "", "Bad Gateway", nil),
			expected: true,
		},
		{
			name:     "404 API error",
			err:      NewAPIError(404, "", "Not Found", nil),
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
		assert.Error(t, ErrPermissionDenied)
		assert.Error(t, ErrInvalidInput)
		assert.Error(t, ErrServer)
	})
}
