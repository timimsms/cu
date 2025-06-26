package errors

import (
	"errors"
	"fmt"
	"net/http"
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

	// ErrInvalidInput indicates invalid user input
	ErrInvalidInput = errors.New("invalid input")

	// ErrConfigNotFound indicates the configuration was not found
	ErrConfigNotFound = errors.New("configuration not found")
)

// APIError represents an error from the ClickUp API
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("API error (%d): %s - %s", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, message string, details ...string) *APIError {
	err := &APIError{
		StatusCode: statusCode,
		Message:    message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// UserError represents a user-friendly error message
type UserError struct {
	Message    string
	Suggestion string
	Err        error
}

func (e *UserError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s\n\nSuggestion: %s", e.Message, e.Suggestion)
	}
	return e.Message
}

func (e *UserError) Unwrap() error {
	return e.Err
}

// NewUserError creates a new user-friendly error
func NewUserError(message, suggestion string, err error) *UserError {
	return &UserError{
		Message:    message,
		Suggestion: suggestion,
		Err:        err,
	}
}

// HandleHTTPError converts HTTP status codes to appropriate errors
func HandleHTTPError(statusCode int, body string) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return NewUserError(
			"Authentication failed",
			"Run 'cu auth login' to authenticate",
			ErrNotAuthenticated,
		)
	case http.StatusForbidden:
		return NewUserError(
			"Access denied",
			"Check your permissions for this resource",
			NewAPIError(statusCode, "Forbidden", body),
		)
	case http.StatusNotFound:
		return NewUserError(
			"Resource not found",
			"Check the ID or name and try again",
			ErrNotFound,
		)
	case http.StatusTooManyRequests:
		return NewUserError(
			"Rate limit exceeded",
			"Please wait a moment and try again",
			ErrRateLimited,
		)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return NewUserError(
			"ClickUp service error",
			"The service is temporarily unavailable. Please try again later",
			NewAPIError(statusCode, "Service Error", body),
		)
	default:
		if statusCode >= 400 {
			return NewAPIError(statusCode, "Request failed", body)
		}
		return nil
	}
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
