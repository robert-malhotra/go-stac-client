package stacclient

import (
	"context"
	"net/http"
	"time"
)

// RetryPolicy decides whether a request should be retried.
type RetryPolicy interface {
	ShouldRetry(resp *http.Response, err error) (bool, time.Duration)
}

// RetryPolicyFunc adapts a function to the RetryPolicy interface.
type RetryPolicyFunc func(resp *http.Response, err error) (bool, time.Duration)

// ShouldRetry implements the RetryPolicy interface.
func (f RetryPolicyFunc) ShouldRetry(resp *http.Response, err error) (bool, time.Duration) {
	return f(resp, err)
}

// DefaultRetryPolicy retries on temporary network errors and server errors with exponential backoff.
var DefaultRetryPolicy RetryPolicy = RetryPolicyFunc(func(resp *http.Response, err error) (bool, time.Duration) {
	switch {
	case err != nil:
		return true, 500 * time.Millisecond
	case resp.StatusCode >= 500:
		return true, 500 * time.Millisecond
	default:
		return false, 0
	}
})

func (c *Client) retry(ctx context.Context, fn func() (*http.Response, error)) (*http.Response, error) {
	policy := c.retryPolicy
	if policy == nil {
		return fn()
	}
	var attempt int
	for {
		resp, err := fn()
		retry, delay := policy.ShouldRetry(resp, err)
		if !retry || ctx.Err() != nil {
			return resp, err
		}
		if resp != nil {
			resp.Body.Close()
		}
		attempt++
		select {
		case <-ctx.Done():
			return resp, ctx.Err()
		case <-time.After(delay * time.Duration(attempt)):
		}
	}
}
