package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
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

// fakeSleeper records requested waits instead of sleeping
type fakeSleeper struct {
	mu    sync.Mutex
	waits []time.Duration
}

func (f *fakeSleeper) sleep(ctx context.Context, d time.Duration) error {
	f.mu.Lock()
	f.waits = append(f.waits, d)
	f.mu.Unlock()
	return ctx.Err()
}

func (f *fakeSleeper) recorded() []time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]time.Duration(nil), f.waits...)
}

// countingBody tracks whether a response body was closed
type countingBody struct {
	io.ReadCloser
	closed *atomic.Int32
}

func (b *countingBody) Close() error {
	b.closed.Add(1)
	return b.ReadCloser.Close()
}

// countingTransport wraps every response body so tests can assert closes
type countingTransport struct {
	base   http.RoundTripper
	opened atomic.Int32
	closed atomic.Int32
}

func (c *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.base.RoundTrip(req)
	if resp != nil && resp.Body != nil {
		c.opened.Add(1)
		resp.Body = &countingBody{ReadCloser: resp.Body, closed: &c.closed}
	}
	return resp, err
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

		sleeper := &fakeSleeper{}
		transport := &retryTransport{base: mock, sleep: sleeper.sleep}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 3, mock.calls, "Should retry twice before success")
		assert.Len(t, sleeper.recorded(), 2, "Should have backoff delays")
	})

	t.Run("retries on 429 rate limit", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 429, body: "rate limited"},
				{statusCode: 200, body: "success"},
			},
		}

		transport := &retryTransport{base: mock, sleep: (&fakeSleeper{}).sleep}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, 2, mock.calls, "Should retry once for rate limit")
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

		transport := &retryTransport{base: mock, sleep: (&fakeSleeper{}).sleep}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)
		assert.Equal(t, 3, mock.calls, "Should stop after 3 attempts")
	})

	t.Run("exponential backoff", func(t *testing.T) {
		mock := &mockRoundTripper{
			responses: []mockResponse{
				{statusCode: 500, body: "error"},
				{statusCode: 500, body: "error"},
				{statusCode: 200, body: "success"},
			},
		}

		sleeper := &fakeSleeper{}
		transport := &retryTransport{base: mock, sleep: sleeper.sleep}
		req, _ := http.NewRequest("GET", "http://example.com", nil)

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		// First retry: 100ms, second retry: 200ms
		assert.Equal(t, []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}, sleeper.recorded())
	})
}

func TestRetryTransportHTTP(t *testing.T) {
	// newServer replies with the given statuses in order, then 200
	newServer := func(t *testing.T, statuses []int, headers http.Header) (*httptest.Server, *atomic.Int32) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := int(calls.Add(1)) - 1
			if n < len(statuses) {
				for k, v := range headers {
					w.Header()[k] = v
				}
				w.WriteHeader(statuses[n])
				_, _ = fmt.Fprintf(w, "attempt %d", n+1)
				return
			}
			_, _ = io.WriteString(w, "ok")
		}))
		t.Cleanup(srv.Close)
		return srv, &calls
	}

	t.Run("closes every intermediate body", func(t *testing.T) {
		srv, calls := newServer(t, []int{429, 503}, http.Header{"Retry-After": {"1"}})
		counter := &countingTransport{base: http.DefaultTransport}
		transport := &retryTransport{base: counter, sleep: (&fakeSleeper{}).sleep}

		req, _ := http.NewRequest("GET", srv.URL, nil)
		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, int32(3), calls.Load())
		assert.Equal(t, int32(3), counter.opened.Load())
		assert.Equal(t, int32(2), counter.closed.Load(), "both discarded bodies must be closed, the returned one left open")
	})

	t.Run("exhausted retries return a readable body", func(t *testing.T) {
		srv, calls := newServer(t, []int{500, 500, 500}, nil)
		counter := &countingTransport{base: http.DefaultTransport}
		transport := &retryTransport{base: counter, sleep: (&fakeSleeper{}).sleep}

		req, _ := http.NewRequest("GET", srv.URL, nil)
		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 500, resp.StatusCode)
		assert.Equal(t, int32(3), calls.Load())
		assert.Equal(t, int32(2), counter.closed.Load())

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "attempt 3", string(data))
	})

	t.Run("honours Retry-After seconds", func(t *testing.T) {
		srv, _ := newServer(t, []int{429}, http.Header{"Retry-After": {"7"}})
		sleeper := &fakeSleeper{}
		transport := &retryTransport{base: http.DefaultTransport, sleep: sleeper.sleep}

		req, _ := http.NewRequest("GET", srv.URL, nil)
		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, []time.Duration{7 * time.Second}, sleeper.recorded())
	})

	t.Run("honours Retry-After HTTP-date", func(t *testing.T) {
		now := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)
		at := now.Add(12 * time.Second).Format(http.TimeFormat)
		srv, _ := newServer(t, []int{429}, http.Header{"Retry-After": {at}})
		sleeper := &fakeSleeper{}
		transport := &retryTransport{
			base:  http.DefaultTransport,
			sleep: sleeper.sleep,
			now:   func() time.Time { return now },
		}

		req, _ := http.NewRequest("GET", srv.URL, nil)
		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, []time.Duration{12 * time.Second}, sleeper.recorded())
	})

	t.Run("caps excessive Retry-After", func(t *testing.T) {
		srv, _ := newServer(t, []int{429}, http.Header{"Retry-After": {"86400"}})
		sleeper := &fakeSleeper{}
		transport := &retryTransport{base: http.DefaultTransport, sleep: sleeper.sleep}

		req, _ := http.NewRequest("GET", srv.URL, nil)
		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, []time.Duration{retryMaxWait}, sleeper.recorded())
	})

	t.Run("context cancel aborts the wait", func(t *testing.T) {
		srv, calls := newServer(t, []int{429, 429, 429}, http.Header{"Retry-After": {"30"}})
		counter := &countingTransport{base: http.DefaultTransport}
		// Real timer-based wait: cancellation must win over a 30s Retry-After
		transport := &retryTransport{base: counter}

		ctx, cancel := context.WithCancel(context.Background())
		req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)

		time.AfterFunc(50*time.Millisecond, cancel)
		start := time.Now()
		resp, err := transport.RoundTrip(req)

		assert.Nil(t, resp)
		require.ErrorIs(t, err, context.Canceled)
		assert.Less(t, time.Since(start), 5*time.Second)
		assert.Equal(t, int32(1), calls.Load())
		assert.Equal(t, int32(1), counter.closed.Load(), "body of the 429 must be closed before waiting")
	})

	t.Run("POST body is resent intact", func(t *testing.T) {
		const payload = `{"name":"task"}`
		var mu sync.Mutex
		var bodies []string
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, _ := io.ReadAll(r.Body)
			mu.Lock()
			bodies = append(bodies, string(data))
			mu.Unlock()
			if calls.Add(1) < 3 {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		transport := &retryTransport{base: http.DefaultTransport, sleep: (&fakeSleeper{}).sleep}
		req, _ := http.NewRequest("POST", srv.URL, bytes.NewBufferString(payload))

		resp, err := transport.RoundTrip(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)
		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, []string{payload, payload, payload}, bodies)
	})
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC)

	tests := []struct {
		name  string
		value string
		want  time.Duration
		ok    bool
	}{
		{"empty", "", 0, false},
		{"seconds", "5", 5 * time.Second, true},
		{"zero", "0", 0, true},
		{"negative", "-1", 0, false},
		{"garbage", "soon", 0, false},
		{"capped seconds", "3600", retryMaxWait, true},
		{"http date", now.Add(30 * time.Second).Format(http.TimeFormat), 30 * time.Second, true},
		{"past http date", now.Add(-time.Minute).Format(http.TimeFormat), 0, true},
		{"capped http date", now.Add(time.Hour).Format(http.TimeFormat), retryMaxWait, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseRetryAfter(tt.value, now)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}
