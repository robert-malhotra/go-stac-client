package client

import (
	"context"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"

	stac "github.com/planetlabs/go-stac"
)

// Middleware manipulates an outgoing *http.Request before it is executed.
type Middleware func(*http.Request) error

// NextHandler determines the next-page URL from a list of STAC links.
// Return nil if there’s no next page, or an error if parsing fails.
type NextHandler func([]*stac.Link) (*url.URL, error)

// ClientOption configures the Client.
type ClientOption func(*Client)

// Client represents a STAC API client
type Client struct {
	baseURL     *url.URL
	httpClient  *http.Client
	nextHandler NextHandler
	middleware  []Middleware
}

// -----------------------------------------------------------------------------
// Client options
// -----------------------------------------------------------------------------

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = client }
}

// WithTimeout sets the HTTP timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithNextHandler configures a custom NextHandler for pagination.
func WithNextHandler(h NextHandler) ClientOption {
	return func(c *Client) { c.nextHandler = h }
}

// WithMiddleware registers one or more request-middleware functions.
func WithMiddleware(mw ...Middleware) ClientOption {
	return func(c *Client) { c.middleware = append(c.middleware, mw...) }
}

// NewClient creates a new STAC client.
func NewClient(baseURL string, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if u.Path != "" && !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	if u.RawPath != "" && !strings.HasSuffix(u.RawPath, "/") {
		u.RawPath += "/"
	}
	c := &Client{
		baseURL:     u,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		nextHandler: DefaultNextHandler,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// DefaultNextHandler looks for the first link with rel="next" and returns its
// Href parsed as a URL. The returned URL may be relative or absolute,
// as specified in the link's Href.
func DefaultNextHandler(links []*stac.Link) (*url.URL, error) {
	nl := findLinkByRel(links, "next")
	if nl == nil {
		return nil, nil // No "next" link found
	}

	if nl.Href == "" {
		return nil, fmt.Errorf("found 'next' link with empty Href")
	}

	parsedNextURL, err := url.Parse(nl.Href)
	if err != nil {
		return nil, fmt.Errorf("invalid 'next' link URL '%s': %w", nl.Href, err)
	}
	return parsedNextURL, nil
}
func findLinkByRel(links []*stac.Link, rel string) *stac.Link {
	for i := range links {
		if links[i].Rel == rel {
			return links[i]
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// iteratePages: generic STAC pagination driver (no type params on methods!)
// -----------------------------------------------------------------------------
//
//   - startPath – relative OR absolute URL for page 1.
//   - decoder   – turns the HTTP body into `(slice-of-T, links)`.
//
// The consumer receives values via an `iter.Seq2[*T,error]` exactly like
// GetItems / GetCollections already expose.

func iteratePages[T any](
	ctx context.Context,
	cli *Client,
	startPath string,
	decoder func(io.Reader) ([]*T, []*stac.Link, error),
) iter.Seq2[*T, error] {

	return func(yield func(*T, error) bool) {
		startURL, err := url.Parse(startPath)
		if err != nil {
			yield(nil, fmt.Errorf("invalid start path %q: %w", startPath, err))
			return
		}

		current := cli.baseURL.ResolveReference(startURL)

		for {
			// --------------------------- HTTP round-trip -------------------
			resp, err := cli.doRequest(ctx, http.MethodGet, current.String(), nil)
			if err != nil {
				if !yield(nil, err) {
					return
				}
				return
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				if !yield(nil, fmt.Errorf("unexpected status %d on %s", resp.StatusCode, current)) {
					return
				}
				return
			}

			// --------------------------- Decode body ----------------------
			items, links, err := decoder(resp.Body)
			resp.Body.Close()
			if err != nil {
				if !yield(nil, fmt.Errorf("error decoding response from %s: %w", current, err)) {
					return
				}
				return
			}

			for _, v := range items {
				if !yield(v, nil) {
					return // consumer stopped
				}
			}

			// --------------------------- Follow “next” --------------------
			next, err := cli.nextHandler(links)
			if err != nil {
				if !yield(nil, fmt.Errorf("error determining next page from %s: %w", current, err)) {
					return
				}
				return
			}
			if next == nil || next.String() == current.String() {
				return // done
			}
			current = cli.baseURL.ResolveReference(next)
		}
	}
}

// -----------------------------------------------------------------------------
// doRequest: one place to build a request, run middleware, and execute it.
// -----------------------------------------------------------------------------
//
// Every endpoint (GetItem, GetCollection, GetItems, GetCollections, Search…)
// should funnel its outbound HTTP calls through this helper so we never repeat
// the boiler-plate middleware loop.
func (c *Client) doRequest(ctx context.Context, method, rawURL string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", rawURL, err)
	}

	if body != nil {
		switch method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if req.Header.Get("Content-Type") == "" {
				req.Header.Set("Content-Type", "application/json")
			}
		}
	}

	// Apply all registered middleware in order.
	for _, mw := range c.middleware {
		if err := mw(req); err != nil {
			return nil, fmt.Errorf("error applying middleware for %s: %w", rawURL, err)
		}
	}

	return c.httpClient.Do(req)
}
