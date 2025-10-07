package stacclient

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidBaseURL is returned when a base URL option is invalid.
	ErrInvalidBaseURL = errors.New("stacclient: invalid base URL")
	// ErrNilHTTPClient indicates a nil HTTP client was provided.
	ErrNilHTTPClient = errors.New("stacclient: http client cannot be nil")
)

// APIError represents a STAC error payload or HTTP failure.
type APIError struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Raw    []byte `json:"-"`
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Title == "" && e.Detail == "" {
		return fmt.Sprintf("stacclient: api error status=%d", e.Status)
	}
	if e.Title != "" && e.Detail != "" {
		return fmt.Sprintf("stacclient: %s (%s)", e.Title, e.Detail)
	}
	if e.Title != "" {
		return fmt.Sprintf("stacclient: %s", e.Title)
	}
	return fmt.Sprintf("stacclient: %s", e.Detail)
}

// Temporary reports whether the error may be retried.
func (e *APIError) Temporary() bool {
	if e == nil {
		return false
	}
	return e.Status >= 500 && e.Status < 600
}
