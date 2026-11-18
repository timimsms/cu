package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Error types
var (
	// ErrNotAuthenticated indicates the user is not authenticated
	ErrNotAuthenticated = errors.New("not authenticated: please run 'cu auth login'")

	// ErrTokenExpired indicates the authentication token has expired
	ErrTokenExpired = errors.New("authentication token expired: please run 'cu auth login' again")

	// ErrInvalidToken indicates the token is invalid
	ErrInvalidToken = errors.New("invalid authentication token")

	// ErrNetworkError indicates a network error occurred
	ErrNetworkError = errors.New("network error")

	// ErrRateLimited indicates the API rate limit was exceeded
	ErrRateLimited = errors.New("rate limit exceeded")

	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrPermissionDenied indicates the token may not access the resource
	ErrPermissionDenied = errors.New("permission denied")

	// ErrInvalidInput indicates invalid user input
	ErrInvalidInput = errors.New("invalid input")

	// ErrServer indicates ClickUp failed to process a valid request (5xx)
	ErrServer = errors.New("ClickUp server error")
)

// APIError is a ClickUp API failure translated into a user-facing message.
// It wraps both a category sentinel (Kind) and the original SDK error (Err),
// so errors.Is(err, ErrNotFound) and errors.As(err, &*clickup.ErrorResponse)
// both keep working.
type APIError struct {
	StatusCode int    `json:"status_code"`
	ECode      string `json:"ecode,omitempty"`   // ClickUp's ECODE, e.g. OAUTH_025
	Message    string `json:"message,omitempty"` // ClickUp's "err" text

	// Resource and ResourceID name what the request was about ("task",
	// "abc123") when it can be derived from the request path.
	Resource   string `json:"resource,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`

	// RetryAfter is how long until the rate limit resets, when known.
	RetryAfter time.Duration `json:"-"`

	Kind error `json:"-"`
	Err  error `json:"-"`
}

// NewAPIError classifies a failed ClickUp response. err is the original
// error and is preserved for unwrapping.
func NewAPIError(statusCode int, ecode, message string, err error) *APIError {
	e := &APIError{StatusCode: statusCode, ECode: ecode, Message: message, Err: err}

	switch {
	case statusCode == http.StatusUnauthorized && isTeamNotAuthorized(ecode, message):
		// ClickUp answers 401 "Team not authorized" when the token is fine but
		// the resource belongs to a workspace it can't see, which is also what
		// a mistyped ID looks like. Treat it as not found, not as a login issue.
		e.Kind = ErrNotFound
	case statusCode == http.StatusUnauthorized:
		e.Kind = ErrNotAuthenticated
	case statusCode == http.StatusForbidden:
		e.Kind = ErrPermissionDenied
	case statusCode == http.StatusNotFound:
		e.Kind = ErrNotFound
	case statusCode == http.StatusTooManyRequests:
		e.Kind = ErrRateLimited
	case statusCode >= 500:
		e.Kind = ErrServer
	case statusCode == http.StatusBadRequest:
		e.Kind = ErrInvalidInput
	}
	return e
}

func isTeamNotAuthorized(ecode, message string) bool {
	return ecode == "OAUTH_027" || strings.Contains(strings.ToLower(message), "team not authorized")
}

func (e *APIError) Error() string {
	return e.summary() + e.detail()
}

// summary is the user-facing explanation of what went wrong.
func (e *APIError) summary() string {
	target := e.target()
	switch {
	case e.Kind == ErrNotFound && e.StatusCode == http.StatusUnauthorized:
		if target == "" {
			target = "resource"
		}
		return target + " not found, or not accessible with the current token"
	case e.Kind == ErrNotFound:
		if target == "" {
			return "resource not found"
		}
		return target + " not found"
	case e.Kind == ErrNotAuthenticated:
		return "not authenticated or token invalid — run `cu auth login`"
	case e.Kind == ErrPermissionDenied:
		if target == "" {
			return "permission denied — check your ClickUp access to this resource"
		}
		return "permission denied for " + target + " — check your ClickUp access to it"
	case e.Kind == ErrRateLimited:
		if e.RetryAfter > 0 {
			return fmt.Sprintf("ClickUp rate limit exceeded — wait %s and try again", e.RetryAfter.Round(time.Second))
		}
		return "ClickUp rate limit exceeded — wait a minute and try again"
	case e.Kind == ErrServer:
		return "ClickUp server error — try again shortly"
	case e.Kind == ErrInvalidInput:
		return "ClickUp rejected the request"
	default:
		return "ClickUp request failed"
	}
}

// detail keeps the raw status, ECODE and ClickUp message visible for
// debugging and bug reports, e.g. " (HTTP 401 OAUTH_025: Token invalid)".
func (e *APIError) detail() string {
	d := fmt.Sprintf("HTTP %d", e.StatusCode)
	if e.ECode != "" {
		d += " " + e.ECode
	}
	if e.Message != "" {
		d += ": " + e.Message
	}
	return " (" + d + ")"
}

func (e *APIError) target() string {
	if e.Resource != "" && e.ResourceID != "" {
		return fmt.Sprintf("%s %q", e.Resource, e.ResourceID)
	}
	return e.Resource
}

// Unwrap exposes both the category sentinel and the original error.
func (e *APIError) Unwrap() []error {
	var errs []error
	if e.Kind != nil {
		errs = append(errs, e.Kind)
	}
	if e.Err != nil {
		errs = append(errs, e.Err)
	}
	return errs
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types
	if errors.Is(err, ErrNetworkError) || errors.Is(err, ErrRateLimited) {
		return true
	}

	// Check for API errors
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusTooManyRequests ||
			apiErr.StatusCode == http.StatusServiceUnavailable ||
			apiErr.StatusCode == http.StatusBadGateway
	}

	return false
}
