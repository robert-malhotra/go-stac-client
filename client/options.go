package stacclient

import (
	"net/http"
	"net/url"
	"time"
)

// Logger represents the minimal logging interface used by the client.
type Logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

// ClientOption configures a Client during construction.
type ClientOption func(*Client) error

// RequestOption configures an outgoing HTTP request at call time.
type RequestOption func(*http.Request) error

// WithBaseURL sets the STAC service base URL.
func WithBaseURL(raw string) ClientOption {
	return func(c *Client) error {
		if raw == "" {
			return ErrInvalidBaseURL
		}
		u, err := url.Parse(raw)
		if err != nil {
			return err
		}
		if !u.IsAbs() {
			return ErrInvalidBaseURL
		}
		c.baseURL = u
		return nil
	}
}

// WithHTTPClient injects a custom http.Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			return ErrNilHTTPClient
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithDefaultHeader registers a header applied to every request.
func WithDefaultHeader(key, value string) ClientOption {
	return func(c *Client) error {
		if key == "" {
			return nil
		}
		if c.defaultHeaders == nil {
			c.defaultHeaders = make(http.Header)
		}
		c.defaultHeaders.Add(key, value)
		return nil
	}
}

// WithRetryPolicy configures the retry behavior for retriable requests.
func WithRetryPolicy(policy RetryPolicy) ClientOption {
	return func(c *Client) error {
		c.retryPolicy = policy
		return nil
	}
}

// WithLogger registers a logger used for request lifecycle events.
func WithLogger(logger Logger) ClientOption {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}

// WithTimeout sets a per-request timeout on the underlying http.Client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		if timeout <= 0 {
			return nil
		}
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
		return nil
	}
}

// Header returns a RequestOption that sets a header value.
func Header(key, value string) RequestOption {
	return func(req *http.Request) error {
		if key == "" {
			return nil
		}
		req.Header.Set(key, value)
		return nil
	}
}

// AddHeader returns a RequestOption that appends to a header value.
func AddHeader(key, value string) RequestOption {
	return func(req *http.Request) error {
		if key == "" {
			return nil
		}
		req.Header.Add(key, value)
		return nil
	}
}
