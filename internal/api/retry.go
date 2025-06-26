package api

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/tim/cu/internal/errors"
)

// retryTransport implements automatic retry with exponential backoff
type retryTransport struct {
	base http.RoundTripper
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
	}

	var resp *http.Response
	var err error
	backoff := 100 * time.Millisecond

	for attempt := 0; attempt < 3; attempt++ {
		// Clone request with body
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		resp, err = t.base.RoundTrip(req)
		
		// Don't retry on success or client errors
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}

		// Check if error is retryable
		if err != nil && !errors.IsRetryable(err) {
			return nil, err
		}

		// Handle rate limiting
		if resp != nil && resp.StatusCode == 429 {
			// Check for Retry-After header
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, parseErr := time.ParseDuration(retryAfter + "s"); parseErr == nil {
					time.Sleep(seconds)
					continue
				}
			}
		}

		// Close response body if exists
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		// Don't sleep on last attempt
		if attempt < 2 {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 5*time.Second {
				backoff = 5 * time.Second
			}
		}
	}

	return resp, err
}