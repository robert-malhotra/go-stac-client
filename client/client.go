package stacclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Client is a reusable STAC API client.
type Client struct {
	httpClient     *http.Client
	baseURL        *url.URL
	defaultHeaders http.Header
	retryPolicy    RetryPolicy
	logger         Logger
}

// New constructs a Client with provided options.
func New(opts ...ClientOption) (*Client, error) {
	c := &Client{
		httpClient:     &http.Client{},
		defaultHeaders: make(http.Header),
		retryPolicy:    DefaultRetryPolicy,
	}
	c.defaultHeaders.Set("Accept", "application/json")
	c.defaultHeaders.Set("Content-Type", "application/json")
	c.defaultHeaders.Set("User-Agent", "go-stac-client/0.1")

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.baseURL == nil {
		return nil, ErrInvalidBaseURL
	}
	if c.httpClient == nil {
		return nil, ErrNilHTTPClient
	}
	return c, nil
}

// Collections returns a service for collection-specific operations.
func (c *Client) Collections() *CollectionService {
	return &CollectionService{client: c}
}

// Items returns a service for item listing and retrieval.
func (c *Client) Items() *ItemService {
	return &ItemService{client: c}
}

// Search returns a service for executing STAC searches.
func (c *Client) Search() *SearchService {
	return &SearchService{client: c}
}

func (c *Client) buildURL(endpoint string, query url.Values) (string, error) {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, endpoint)
	if strings.HasSuffix(endpoint, "/") && !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u.String(), nil
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, query url.Values, body any, opts []RequestOption) (*http.Request, error) {
	urlStr, err := c.buildURL(endpoint, query)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if body != nil {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
		reader = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, reader)
	if err != nil {
		return nil, err
	}

	for key, values := range c.defaultHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(req); err != nil {
			return nil, err
		}
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if c.logger != nil {
		c.logger.Debugf("stacclient: %s %s", req.Method, req.URL)
	}

	resp, err := c.retry(ctx, func() (*http.Response, error) {
		return c.httpClient.Do(req)
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	defer resp.Body.Close()
	data, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if readErr != nil {
		return nil, readErr
	}

	apiErr := &APIError{Status: resp.StatusCode, Raw: data}
	if err := json.Unmarshal(data, apiErr); err != nil {
		// Fallback to plain message.
		apiErr.Detail = string(data)
	}
	if c.logger != nil {
		c.logger.Errorf("stacclient: request failed status=%d", resp.StatusCode)
	}
	return nil, apiErr
}

func (c *Client) doJSON(ctx context.Context, method, endpoint string, query url.Values, body any, out any, opts []RequestOption) error {
	req, err := c.newRequest(ctx, method, endpoint, query, body, opts)
	if err != nil {
		return err
	}

	resp, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}

func cloneValues(values url.Values) url.Values {
	if len(values) == 0 {
		return nil
	}
	cp := make(url.Values, len(values))
	for key, v := range values {
		dst := make([]string, len(v))
		copy(dst, v)
		cp[key] = dst
	}
	return cp
}
