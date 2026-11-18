package api

import (
	"context"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timimsms/cu/internal/errors"
)

// errorServer replies to every request with status, the given headers and a
// ClickUp-style error body.
func errorServer(t *testing.T, status int, body string, headers map[string]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestHandleError_StatusCodes(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		headers map[string]string
		kind    error
		want    string
	}{
		{
			name:   "401 token invalid",
			status: http.StatusUnauthorized,
			body:   `{"err":"Token invalid","ECODE":"OAUTH_025"}`,
			kind:   errors.ErrNotAuthenticated,
			want:   "not authenticated or token invalid — run `cu auth login` (HTTP 401 OAUTH_025: Token invalid)",
		},
		{
			name:   "401 team not authorized",
			status: http.StatusUnauthorized,
			body:   `{"err":"Team not authorized","ECODE":"OAUTH_027"}`,
			kind:   errors.ErrNotFound,
			want:   `task "abc123" not found, or not accessible with the current token (HTTP 401 OAUTH_027: Team not authorized)`,
		},
		{
			name:   "403",
			status: http.StatusForbidden,
			body:   `{"err":"Forbidden","ECODE":"ACCESS_001"}`,
			kind:   errors.ErrPermissionDenied,
			want:   `permission denied for task "abc123" — check your ClickUp access to it (HTTP 403 ACCESS_001: Forbidden)`,
		},
		{
			name:   "404",
			status: http.StatusNotFound,
			body:   `{"err":"Task not found","ECODE":"ITEM_013"}`,
			kind:   errors.ErrNotFound,
			want:   `task "abc123" not found (HTTP 404 ITEM_013: Task not found)`,
		},
		{
			name:    "429 with Retry-After",
			status:  http.StatusTooManyRequests,
			body:    `{"err":"Rate limit reached","ECODE":"APP_002"}`,
			headers: map[string]string{"Retry-After": "30"},
			kind:    errors.ErrRateLimited,
			want:    "ClickUp rate limit exceeded — wait 30s and try again (HTTP 429 APP_002: Rate limit reached)",
		},
		{
			name:   "429 without hints",
			status: http.StatusTooManyRequests,
			body:   `{"err":"Rate limit reached","ECODE":"APP_002"}`,
			kind:   errors.ErrRateLimited,
			want:   "ClickUp rate limit exceeded — wait a minute and try again (HTTP 429 APP_002: Rate limit reached)",
		},
		{
			name:   "500",
			status: http.StatusInternalServerError,
			body:   `{"err":"Internal Server Error","ECODE":"SHARD_001"}`,
			kind:   errors.ErrServer,
			want:   "ClickUp server error — try again shortly (HTTP 500 SHARD_001: Internal Server Error)",
		},
		{
			name:   "503 with non-JSON body",
			status: http.StatusServiceUnavailable,
			body:   `<html>upstream unavailable</html>`,
			kind:   errors.ErrServer,
			want:   "ClickUp server error — try again shortly (HTTP 503)",
		},
		{
			name:   "400",
			status: http.StatusBadRequest,
			body:   `{"err":"Status does not exist","ECODE":"ITEM_102"}`,
			kind:   errors.ErrInvalidInput,
			want:   "ClickUp rejected the request (HTTP 400 ITEM_102: Status does not exist)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := errorServer(t, tt.status, tt.body, tt.headers)

			_, err := newTestClient(t, server).GetTask(context.Background(), "abc123")
			require.Error(t, err)

			assert.Equal(t, tt.want, err.Error())
			assert.ErrorIs(t, err, tt.kind)

			var apiErr *errors.APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tt.status, apiErr.StatusCode)

			var sdkErr *clickup.ErrorResponse
			assert.ErrorAs(t, err, &sdkErr, "the SDK error must stay reachable")
		})
	}
}

func TestHandleError_RateLimitError(t *testing.T) {
	reset := time.Now().Add(45 * time.Second).Unix()
	server := errorServer(t, http.StatusTooManyRequests, `{"err":"Rate limit reached","ECODE":"APP_002"}`, map[string]string{
		"X-RateLimit-Limit":     "100",
		"X-RateLimit-Remaining": "0",
		"X-RateLimit-Reset":     strconv.FormatInt(reset, 10),
	})
	client := newTestClient(t, server)

	_, err := client.GetTask(context.Background(), "abc123")
	require.Error(t, err)

	var sdkErr *clickup.RateLimitError
	require.ErrorAs(t, err, &sdkErr)
	assert.ErrorIs(t, err, errors.ErrRateLimited)
	assert.Contains(t, err.Error(), "ClickUp rate limit exceeded — wait 4")

	// The SDK now refuses further requests locally with a synthetic 403;
	// that must still read as a rate limit, not a permission problem.
	_, err = client.GetTask(context.Background(), "abc123")
	require.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrRateLimited)
	assert.NotErrorIs(t, err, errors.ErrPermissionDenied)
}

func TestHandleError_ResourceContext(t *testing.T) {
	ctx := context.Background()
	body := `{"err":"Not found","ECODE":"X_404"}`

	tests := []struct {
		name string
		call func(*Client) error
		want string
	}{
		{
			name: "list tasks",
			call: func(c *Client) error { _, err := c.GetTasks(ctx, "L1", &TaskQueryOptions{}); return err },
			want: `list "L1" not found`,
		},
		{
			name: "folder lists",
			call: func(c *Client) error { _, err := c.GetLists(ctx, "F1"); return err },
			want: `folder "F1" not found`,
		},
		{
			name: "team tasks",
			call: func(c *Client) error { _, err := c.SearchTeamTasks(ctx, "T1", &TeamTaskQuery{}); return err },
			want: `workspace "T1" not found`,
		},
		{
			name: "custom field on task",
			call: func(c *Client) error { return c.RemoveCustomFieldValue(ctx, "abc", "fld") },
			want: `task "abc" not found`,
		},
		{
			name: "workspaces",
			call: func(c *Client) error { _, err := c.GetWorkspaces(ctx); return err },
			want: `workspace not found`,
		},
		{
			name: "authorized user",
			call: func(c *Client) error { _, err := c.GetCurrentUser(ctx); return err },
			want: `resource not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := errorServer(t, http.StatusNotFound, body, nil)

			// Mirror production, where the SDK's base URL has an /api/v2/ prefix.
			client := newTestClient(t, server)
			base, err := url.Parse(server.URL + "/api/v2/")
			require.NoError(t, err)
			client.client.BaseURL = base

			err = tt.call(client)
			require.Error(t, err)
			assert.Equal(t, tt.want+" (HTTP 404 X_404: Not found)", err.Error())
		})
	}
}

func TestHandleError_PassThrough(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.NoError(t, (&Client{}).handleError(nil))
	})

	t.Run("network error", func(t *testing.T) {
		server := httptest.NewServer(http.NotFoundHandler())
		client := newTestClient(t, server)
		server.Close()

		_, err := client.GetTask(context.Background(), "abc123")
		require.Error(t, err)

		var apiErr *errors.APIError
		assert.False(t, stderrors.As(err, &apiErr), "transport failures are not API errors")
		var urlErr *url.Error
		assert.ErrorAs(t, err, &urlErr)
	})

	t.Run("context canceled", func(t *testing.T) {
		err := fmt.Errorf("request: %w", context.Canceled)
		assert.Same(t, err, (&Client{}).handleError(err))
	})

	t.Run("already typed", func(t *testing.T) {
		typed := errors.NewAPIError(http.StatusNotFound, "", "", nil)
		assert.Same(t, typed, (&Client{}).handleError(typed))
	})
}
