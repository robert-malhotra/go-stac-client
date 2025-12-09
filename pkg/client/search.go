package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strings"

	"github.com/robert-malhotra/go-stac-client/pkg/stac"
)

// -----------------------------------------------------------------------------
// Domain model & search types (assumed unchanged from original)
// -----------------------------------------------------------------------------

type SearchParams struct {
	Collections []string       `json:"collections,omitempty"`
	Bbox        []float64      `json:"bbox,omitempty"`
	Datetime    string         `json:"datetime,omitempty"`
	Query       map[string]any `json:"query,omitempty"`
	Limit       int            `json:"limit,omitempty"`
	SortBy      []SortField    `json:"sortby,omitempty"`
	Fields      *FieldsFilter  `json:"fields,omitempty"`
}

type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" or "desc"
}

type FieldsFilter struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

type Error struct {
	Code        int    `json:"code"` // HTTP status code
	Description string `json:"description"`
	Type        string `json:"type,omitempty"` // Specific error type if provided by API
}

// SearchSimple performs a GET-based STAC search using URL query parameters.
func (c *Client) SearchSimple(ctx context.Context, params SearchParams) iter.Seq2[*stac.Item, error] {
	// Build query parameters
	q := url.Values{}
	for _, coll := range params.Collections {
		q.Add("collections", coll)
	}

	var marshalErr error
	if len(params.Bbox) >= 4 && len(params.Bbox)%2 == 0 {
		coords := make([]string, len(params.Bbox))
		for i, v := range params.Bbox {
			coords[i] = fmt.Sprintf("%g", v)
		}
		q.Set("bbox", strings.Join(coords, ","))
	}
	if params.Datetime != "" {
		q.Set("datetime", params.Datetime)
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if len(params.SortBy) > 0 {
		var parts []string
		for _, s := range params.SortBy {
			dir := strings.ToLower(s.Direction)
			if dir != "asc" && dir != "desc" {
				dir = "desc"
			}
			parts = append(parts, fmt.Sprintf("%s:%s", s.Field, dir))
		}
		q.Set("sortby", strings.Join(parts, ","))
	}
	if params.Query != nil {
		if queryJSON, err := json.Marshal(params.Query); err == nil {
			q.Set("query", string(queryJSON))
		} else if marshalErr == nil {
			marshalErr = fmt.Errorf("error encoding query parameters: %w", err)
		}
	}
	if params.Fields != nil {
		if fieldsJSON, err := json.Marshal(params.Fields); err == nil {
			q.Set("fields", string(fieldsJSON))
		} else if marshalErr == nil {
			marshalErr = fmt.Errorf("error encoding fields parameters: %w", err)
		}
	}
	if marshalErr != nil {
		return func(y func(*stac.Item, error) bool) {
			y(nil, marshalErr)
		}
	}

	startURL := &url.URL{Path: "search", RawQuery: q.Encode()}

	return iteratePages[stac.Item](ctx, c, startURL.String(),
		func(r io.Reader) ([]*stac.Item, []*stac.Link, error) {
			var page struct {
				Features []*stac.Item `json:"features"`
				Links    []*stac.Link `json:"links"`
			}
			err := json.NewDecoder(r).Decode(&page)
			return page.Features, page.Links, err
		})
}

// SearchCQL2 performs a POST-based STAC search using the provided SearchParams as JSON payload.
func (c *Client) SearchCQL2(ctx context.Context, params SearchParams) iter.Seq2[*stac.Item, error] {
	// Marshal the search parameters into JSON
	bodyBytes, err := json.Marshal(params)
	if err != nil {
		// Return an iterator that immediately yields the error
		return func(yield func(*stac.Item, error) bool) {
			yield(nil, fmt.Errorf("error marshalling search parameters: %w", err))
		}
	}

	return func(yield func(*stac.Item, error) bool) {
		current := c.baseURL.ResolveReference(&url.URL{Path: "search"})
		usePOST := true

		for {
			var (
				method = http.MethodGet
				body   io.Reader
			)
			if usePOST {
				method = http.MethodPost
				body = bytes.NewReader(bodyBytes)
			}

			resp, err := c.doRequest(ctx, method, current.String(), body)
			if err != nil {
				yield(nil, err)
				return
			}
			if resp.StatusCode != http.StatusOK {
				defer resp.Body.Close()
				var apiErr Error
				if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
					yield(nil, fmt.Errorf("unexpected status %d on %s", resp.StatusCode, current))
					return
				}
				if apiErr.Code == 0 {
					apiErr.Code = resp.StatusCode
				}
				yield(nil, fmt.Errorf("search error: %s (code %d, type %s)", apiErr.Description, apiErr.Code, apiErr.Type))
				return
			}

			var page struct {
				Features []*stac.Item `json:"features"`
				Links    []*stac.Link `json:"links"`
			}
			err = json.NewDecoder(resp.Body).Decode(&page)
			resp.Body.Close()
			if err != nil {
				yield(nil, fmt.Errorf("error decoding response from %s: %w", current, err))
				return
			}

			for _, it := range page.Features {
				if !yield(it, nil) {
					return
				}
			}

			nextURL, err := c.nextHandler(page.Links)
			if err != nil {
				yield(nil, fmt.Errorf("error determining next page from %s: %w", current, err))
				return
			}
			if nextURL == nil {
				return
			}
			current = c.baseURL.ResolveReference(nextURL)
			usePOST = false
		}
	}
}
