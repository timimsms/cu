package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient builds a Client whose SDK client talks to server instead of
// the real ClickUp API.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	sdk := clickup.NewClient(server.Client(), "test-token")
	baseURL, err := url.Parse(server.URL + "/")
	require.NoError(t, err)
	sdk.BaseURL = baseURL

	return &Client{
		client:      sdk,
		rateLimiter: NewRateLimiter(100, time.Minute),
	}
}

func TestSetTaskArchived(t *testing.T) {
	tests := []struct {
		name     string
		archived bool
	}{
		{name: "archive", archived: true},
		{name: "undo", archived: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody map[string]any

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &gotBody)

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":"abc123","name":"Retired task"}`))
			}))
			defer server.Close()

			task, err := newTestClient(t, server).SetTaskArchived(context.Background(), "abc123", tt.archived)
			require.NoError(t, err)

			assert.Equal(t, http.MethodPut, gotMethod)
			assert.Equal(t, "/task/abc123", gotPath)
			assert.Equal(t, map[string]any{"archived": tt.archived}, gotBody)
			assert.Equal(t, "abc123", task.ID)
			assert.Equal(t, "Retired task", task.Name)
		})
	}
}

func TestSetTaskArchived_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"err":"Task not found","ECODE":"ITEM_013"}`))
	}))
	defer server.Close()

	task, err := newTestClient(t, server).SetTaskArchived(context.Background(), "missing", true)
	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestSetTaskArchived_EscapesTaskID(t *testing.T) {
	var gotRawPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawPath = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"id":"a/b"}`))
	}))
	defer server.Close()

	_, err := newTestClient(t, server).SetTaskArchived(context.Background(), "a/b", true)
	require.NoError(t, err)
	assert.Equal(t, "/task/a%2Fb", gotRawPath)
}

func TestSetTaskArchived_RequiresTaskID(t *testing.T) {
	_, err := (&Client{}).SetTaskArchived(context.Background(), "", true)
	assert.Error(t, err)
}
