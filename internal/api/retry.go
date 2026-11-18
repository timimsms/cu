package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/timimsms/cu/internal/errors"
)

const (
	retryMaxAttempts  = 3
	retryBaseBackoff  = 100 * time.Millisecond
	retryMaxBackoff   = 5 * time.Second
	retryMaxWait      = 60 * time.Second // cap on server-requested Retry-After waits
	retryDrainMaxSize = 64 << 10         // bytes drained from a discarded body for connection reuse
)

// retryTransport implements automatic retry with exponential backoff
type retryTransport struct {
	base http.RoundTripper

	// sleep and now are injectable for tests; nil means real time.
	sleep func(ctx context.Context, d time.Duration) error
	now   func() time.Time
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Buffer the body once so every attempt can resend it intact
	var body []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		body, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	backoff := retryBaseBackoff

	for attempt := 0; ; attempt++ {
		// Clone so the caller's request is never mutated
		r := req.Clone(req.Context())
		if body != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		}

		resp, err := t.base.RoundTrip(r)

		// Don't retry on success or client errors
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Check if error is retryable
		if err != nil && !errors.IsRetryable(err) {
			return nil, err
		}

		// Out of attempts: hand back the last response with its body still open
		if attempt == retryMaxAttempts-1 {
			return resp, err
		}

		// Prefer the server's Retry-After over our own backoff
		wait := backoff
		if resp != nil {
			if d, ok := parseRetryAfter(resp.Header.Get("Retry-After"), t.clock()); ok {
				wait = d
			}
		}

		drainAndClose(resp)

		if err := t.wait(req.Context(), wait); err != nil {
			return nil, err
		}

		backoff *= 2
		if backoff > retryMaxBackoff {
			backoff = retryMaxBackoff
		}
	}
}

func (t *retryTransport) clock() time.Time {
	if t.now != nil {
		return t.now()
	}
	return time.Now()
}

func (t *retryTransport) wait(ctx context.Context, d time.Duration) error {
	if t.sleep != nil {
		return t.sleep(ctx, d)
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// parseRetryAfter reads a Retry-After value in either delay-seconds or
// HTTP-date form (RFC 7231 §7.1.3), capped at retryMaxWait.
func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	var d time.Duration
	if secs, err := strconv.ParseInt(value, 10, 64); err == nil {
		if secs < 0 {
			return 0, false
		}
		if secs > int64(retryMaxWait/time.Second) {
			return retryMaxWait, true
		}
		d = time.Duration(secs) * time.Second
	} else if at, err := http.ParseTime(value); err == nil {
		d = at.Sub(now)
		if d < 0 {
			d = 0
		}
	} else {
		return 0, false
	}

	if d > retryMaxWait {
		d = retryMaxWait
	}
	return d, true
}

// drainAndClose discards a response we won't return so its connection can be reused.
func drainAndClose(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, retryDrainMaxSize))
	_ = resp.Body.Close()
}
