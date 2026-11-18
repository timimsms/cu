package api

import (
	stderrors "errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/timimsms/cu/internal/errors"
)

// resourceNames maps the first segment of a ClickUp API path to the noun
// used in error messages. ClickUp calls workspaces "teams" in its API.
var resourceNames = map[string]string{
	"task":    "task",
	"list":    "list",
	"folder":  "folder",
	"space":   "space",
	"team":    "workspace",
	"comment": "comment",
	"view":    "view",
	"goal":    "goal",
}

// handleError converts SDK errors for failed ClickUp responses into typed,
// user-friendly *errors.APIError values. Network and context errors, and
// errors that are already typed, pass through unchanged.
func (c *Client) handleError(err error) error {
	if err == nil {
		return nil
	}

	var typed *errors.APIError
	if stderrors.As(err, &typed) {
		return err
	}

	var (
		resp       *http.Response
		status     int
		ecode, msg string
		retryAfter time.Duration
	)

	var rateErr *clickup.RateLimitError
	var abuseErr *clickup.AbuseRateLimitError
	var respErr *clickup.ErrorResponse
	switch {
	case stderrors.As(err, &rateErr):
		// The SDK also returns this without a request when it already knows
		// the limit is exhausted, using a synthetic 403 response.
		resp, status, msg = rateErr.Response, http.StatusTooManyRequests, rateErr.Message
		if reset := rateErr.Rate.Reset.Time; !reset.IsZero() {
			retryAfter = time.Until(reset)
		}
	case stderrors.As(err, &abuseErr):
		resp, status, msg = abuseErr.Response, http.StatusTooManyRequests, abuseErr.Message
	case stderrors.As(err, &respErr) && respErr.Response != nil:
		resp, status = respErr.Response, respErr.Response.StatusCode
		ecode, msg = respErr.ECode, respErr.Err
		if msg == "" {
			msg = respErr.Message
		}
	default:
		return err
	}

	apiErr := errors.NewAPIError(status, ecode, msg, err)
	if resp != nil {
		if resp.Request != nil && resp.Request.URL != nil {
			apiErr.Resource, apiErr.ResourceID = c.resourceFromPath(resp.Request.URL.Path)
		}
		if retryAfter <= 0 && status == http.StatusTooManyRequests {
			retryAfter = retryAfterFromHeaders(resp.Header)
		}
	}
	apiErr.RetryAfter = retryAfter
	return apiErr
}

// resourceFromPath extracts the resource noun and ID from a request path such
// as /api/v2/task/abc123/field/xyz, yielding ("task", "abc123").
func (c *Client) resourceFromPath(path string) (string, string) {
	if c.client != nil && c.client.BaseURL != nil {
		path = strings.TrimPrefix(path, c.client.BaseURL.Path)
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	name, ok := resourceNames[segments[0]]
	if !ok {
		return "", ""
	}
	if len(segments) < 2 {
		return name, ""
	}
	return name, segments[1]
}

// retryAfterFromHeaders reads the wait time from Retry-After (seconds) or
// ClickUp's X-RateLimit-Reset (unix seconds).
func retryAfterFromHeaders(h http.Header) time.Duration {
	if s, err := strconv.Atoi(h.Get("Retry-After")); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	if ts, err := strconv.ParseInt(h.Get("X-RateLimit-Reset"), 10, 64); err == nil {
		if d := time.Until(time.Unix(ts, 0)); d > 0 {
			return d
		}
	}
	return 0
}
