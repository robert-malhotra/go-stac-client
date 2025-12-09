package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/robert-malhotra/go-stac-client/pkg/stac"
)

// Middleware manipulates an outgoing *http.Request before it is executed.
// The context is provided for cancellation and to support auth implementations
// that may need to perform async operations (e.g., token refresh).
type Middleware func(context.Context, *http.Request) error

// NextHandler determines the next-page URL from a list of STAC links.
// Return nil if there's no next page, or an error if parsing fails.
type NextHandler func([]*stac.Link) (*url.URL, error)

// PageResponse holds the decoded response from a paginated API call.
type PageResponse[T any] struct {
	Items   []*T
	Links   []*stac.Link
	Cursor  string   // For cursor-based pagination (e.g., ICEYE)
	NextURL *url.URL // Pre-computed next URL (optional, takes precedence over Links)
}

// PageDecoder decodes a paginated response body into items and pagination info.
type PageDecoder[T any] func(r io.Reader) (*PageResponse[T], error)

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
// Page Decoders - helpers for different API response formats
// -----------------------------------------------------------------------------

// DefaultItemDecoder creates a decoder for standard STAC item responses.
// Standard STAC APIs return {"features": [...], "links": [...]}.
func DefaultItemDecoder() PageDecoder[stac.Item] {
	return func(r io.Reader) (*PageResponse[stac.Item], error) {
		var page struct {
			Features []*stac.Item `json:"features"`
			Links    []*stac.Link `json:"links"`
		}
		if err := json.NewDecoder(r).Decode(&page); err != nil {
			return nil, err
		}
		return &PageResponse[stac.Item]{Items: page.Features, Links: page.Links}, nil
	}
}

// DefaultCollectionDecoder creates a decoder for standard STAC collection responses.
// Standard STAC APIs return {"collections": [...], "links": [...]}.
func DefaultCollectionDecoder() PageDecoder[stac.Collection] {
	return func(r io.Reader) (*PageResponse[stac.Collection], error) {
		var page struct {
			Collections []*stac.Collection `json:"collections"`
			Links       []*stac.Link       `json:"links"`
		}
		if err := json.NewDecoder(r).Decode(&page); err != nil {
			return nil, err
		}
		return &PageResponse[stac.Collection]{Items: page.Collections, Links: page.Links}, nil
	}
}

// CursorItemDecoder creates a decoder for cursor-based pagination APIs like ICEYE.
// These APIs return {"data": [...], "cursor": "..."} instead of STAC-standard format.
// The nextURLTemplate should contain "%s" where the cursor value will be substituted.
// Example: "/catalog/v2/items?cursor=%s"
func CursorItemDecoder(itemsField, cursorField, nextURLTemplate string) PageDecoder[stac.Item] {
	return func(r io.Reader) (*PageResponse[stac.Item], error) {
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r).Decode(&raw); err != nil {
			return nil, err
		}

		resp := &PageResponse[stac.Item]{}

		// Decode items from the specified field
		if data, ok := raw[itemsField]; ok {
			if err := json.Unmarshal(data, &resp.Items); err != nil {
				return nil, fmt.Errorf("decode %s: %w", itemsField, err)
			}
		}

		// Decode cursor from the specified field
		if c, ok := raw[cursorField]; ok {
			if err := json.Unmarshal(c, &resp.Cursor); err != nil {
				return nil, fmt.Errorf("decode %s: %w", cursorField, err)
			}
		}

		// Build next URL from cursor if present
		if resp.Cursor != "" && nextURLTemplate != "" {
			nextURL, err := url.Parse(fmt.Sprintf(nextURLTemplate, url.QueryEscape(resp.Cursor)))
			if err != nil {
				return nil, fmt.Errorf("build next URL: %w", err)
			}
			resp.NextURL = nextURL
		}

		return resp, nil
	}
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
	// Wrap old-style decoder into new PageDecoder format
	pageDecoder := func(r io.Reader) (*PageResponse[T], error) {
		items, links, err := decoder(r)
		if err != nil {
			return nil, err
		}
		return &PageResponse[T]{Items: items, Links: links}, nil
	}
	return iteratePagesWithDecoder(ctx, cli, startPath, pageDecoder)
}

// iteratePagesWithDecoder is the generic pagination driver that supports both
// link-based (standard STAC) and cursor-based (e.g., ICEYE) pagination.
//
//   - startPath – relative OR absolute URL for page 1.
//   - decoder   – PageDecoder that parses response and extracts pagination info.
//
// The decoder can return either:
//   - Links for standard STAC pagination (uses client's NextHandler)
//   - NextURL for pre-computed next page URL (takes precedence over Links)
//   - Cursor for cursor-based APIs (decoder should build NextURL from cursor)
func iteratePagesWithDecoder[T any](
	ctx context.Context,
	cli *Client,
	startPath string,
	decoder PageDecoder[T],
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
			page, err := decoder(resp.Body)
			resp.Body.Close()
			if err != nil {
				if !yield(nil, fmt.Errorf("error decoding response from %s: %w", current, err)) {
					return
				}
				return
			}

			for _, v := range page.Items {
				if !yield(v, nil) {
					return // consumer stopped
				}
			}

			// --------------------------- Follow "next" --------------------
			// Priority: NextURL > Links (via nextHandler)
			var next *url.URL
			if page.NextURL != nil {
				next = page.NextURL
			} else if len(page.Links) > 0 {
				next, err = cli.nextHandler(page.Links)
				if err != nil {
					if !yield(nil, fmt.Errorf("error determining next page from %s: %w", current, err)) {
						return
					}
					return
				}
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

	// Apply all registered middleware in order.
	for _, mw := range c.middleware {
		if err := mw(ctx, req); err != nil {
			return nil, fmt.Errorf("error applying middleware for %s: %w", rawURL, err)
		}
	}

	return c.httpClient.Do(req)
}
