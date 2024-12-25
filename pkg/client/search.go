package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	stac "github.com/planetlabs/go-stac"
)

// SearchParams represents query parameters for searching items
type SearchParams struct {
	Limit       *int       `json:"limit,omitempty"`
	Collections []string   `json:"collections,omitempty"`
	BBox        []float64  `json:"bbox,omitempty"`
	DateTime    *time.Time `json:"datetime,omitempty"`
	DateEnd     *time.Time `json:"datetime_end,omitempty"`
	Intersects  *Geometry  `json:"intersects,omitempty"`
	IDs         []string   `json:"ids,omitempty"`
	Fields      []string   `json:"fields,omitempty"`
	SortBy      []string   `json:"sortby,omitempty"`
	Filter      string     `json:"filter,omitempty"`
}

// buildSearchQuery converts SearchParams to url.Values
func (p SearchParams) buildSearchQuery() url.Values {
	values := url.Values{}

	if p.Limit != nil {
		values.Set("limit", fmt.Sprintf("%d", *p.Limit))
	}

	if len(p.Collections) > 0 {
		values.Set("collections", strings.Join(p.Collections, ","))
	}

	if len(p.BBox) > 0 {
		values.Set("bbox", joinFloat64(p.BBox))
	}

	if p.DateTime != nil {
		if p.DateEnd != nil {
			values.Set("datetime", fmt.Sprintf("%s/%s",
				p.DateTime.Format(time.RFC3339),
				p.DateEnd.Format(time.RFC3339)))
		} else {
			values.Set("datetime", p.DateTime.Format(time.RFC3339))
		}
	}

	if p.Intersects != nil {
		intersects, err := json.Marshal(p.Intersects)
		if err == nil {
			values.Set("intersects", string(intersects))
		}
	}

	if len(p.IDs) > 0 {
		values.Set("ids", strings.Join(p.IDs, ","))
	}

	for _, field := range p.Fields {
		values.Add("fields", field)
	}

	for _, sort := range p.SortBy {
		values.Add("sortby", sort)
	}

	if p.Filter != "" {
		values.Set("filter", p.Filter)
	}

	return values
}

// Search searches for items across all or specified collections
func (c *Client) Search(ctx context.Context, params *SearchParams) (*stac.ItemsList, error) {
	var query url.Values
	if params != nil {
		query = params.buildSearchQuery()
	}

	u := c.buildURL(path.Join("/search"), query)

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return c.handleItemsRequest(req)
}

// GetNextSearchResults retrieves the next page of search results
func (c *Client) GetNextSearchResults(ctx context.Context, items *stac.ItemsList) (*stac.ItemsList, error) {
	for _, l := range items.Links {
		if strings.ToLower(l.Rel) == "next" {
			u, err := url.Parse(l.Href)
			if err != nil {
				return nil, err
			}
			req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
			if err != nil {
				return nil, fmt.Errorf("error creating request: %w", err)
			}
			return c.handleItemsRequest(req)
		}
	}

	return nil, fmt.Errorf("no link with `next` rel found")
}
