package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRoundTripper helps test retry logic
type mockRoundTripper struct {
	responses []mockResponse
	calls     int
}

type mockResponse struct {
	statusCode int
	body       string
	err        error
	headers    map[string]string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.calls >= len(m.responses) {
		return nil, fmt.Errorf("no more mock responses")
	}

	resp := m.responses[m.calls]
	m.calls++

	if resp.err != nil {
		return nil, resp.err
	}

	r := &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(resp.body)),
		Header:     make(http.Header),
	}

	for k, v := range resp.headers {
		r.Header.Set(k, v)
	}

	return r, nil
}

func TestRetryTransport(t *testing.T) {
	t.Run("successful request - no retry", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 200, body: "success"},
			},
		}

		transport := &retryTransport{base: mock}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 1, mock.calls, "Should only call once for success")
	})

	t.Run("retries on 500 error", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 500, body: "server error"},
				{statusCode: 500, body: "server error"},
				{statusCode: 200, body: "success"},
			},
		}

		transport := &retryTransport{base: mock}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		start := time.Now()
		resp, err := transport.RoundTrip(req)
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 3, mock.calls, "Should retry twice before success")
		assert.True(t, elapsed >= 300*time.Millisecond, "Should have backoff delays")
	})

	t.Run("retries on 429 rate limit", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 429, body: "rate limited"},
				{statusCode: 200, body: "success"},
			},
		}

		transport := &retryTransport{base: mock}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 2, mock.calls, "Should retry once for rate limit")
	})

	t.Run("respects Retry-After header", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{
					statusCode: 429,
					body:       "rate limited",
					headers:    map[string]string{"Retry-After": "1"},
				},
				{statusCode: 200, body: "success"},
			},
		}

		transport := &retryTransport{base: mock}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		start := time.Now()
		resp, err := transport.RoundTrip(req)
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.True(t, elapsed >= 1*time.Second, "Should respect Retry-After header")
	})

	t.Run("does not retry client errors", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 404, body: "not found"},
			},
		}

		transport := &retryTransport{base: mock}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode)
		assert.Equal(t, 1, mock.calls, "Should not retry client errors")
	})

	t.Run("gives up after max retries", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 500, body: "error"},
				{statusCode: 500, body: "error"},
				{statusCode: 500, body: "error"},
			},
		}

		transport := &retryTransport{base: mock}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)
		assert.Equal(t, 3, mock.calls, "Should stop after 3 attempts")
	})

	t.Run("handles request with body", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 500, body: "error"},
				{statusCode: 200, body: "success"},
			},
		}

		transport := &retryTransport{base: mock}
		body := bytes.NewBufferString("request body")
		req, _ := http.NewRequest("POST", "http://example.com", body)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 2, mock.calls, "Should retry with body")
	})

	t.Run("exponential backoff", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 500, body: "error"},
				{statusCode: 500, body: "error"},
				{statusCode: 200, body: "success"},
			},
		}

		transport := &retryTransport{base: mock}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		start := time.Now()
		resp, err := transport.RoundTrip(req)
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		// First retry: 100ms, Second retry: 200ms, Total: 300ms minimum
		assert.True(t, elapsed >= 300*time.Millisecond, "Should use exponential backoff")
		assert.True(t, elapsed < 500*time.Millisecond, "Should not exceed expected backoff")
	})
}