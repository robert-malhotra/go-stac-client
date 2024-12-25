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

// ItemsParams represents query parameters for listing items
type ItemsParams struct {
	Limit      *int       `json:"limit,omitempty"`
	BBox       []float64  `json:"bbox,omitempty"`
	DateTime   *time.Time `json:"datetime,omitempty"`
	DateEnd    *time.Time `json:"datetime_end,omitempty"`
	Intersects *Geometry  `json:"intersects,omitempty"`
	IDs        []string   `json:"ids,omitempty"`
	Fields     []string   `json:"fields,omitempty"`
	SortBy     []string   `json:"sortby,omitempty"`
	Filter     string     `json:"filter,omitempty"`
	Next       string     `json:"next,omitempty"`
}

// Geometry represents a GeoJSON geometry
type Geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// buildItemsQuery converts ItemsParams to url.Values
func (p ItemsParams) buildItemsQuery() url.Values {
	values := url.Values{}

	if p.Limit != nil {
		values.Set("limit", fmt.Sprintf("%d", *p.Limit))
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

// joinFloat64 converts a slice of float64 to a comma-separated string
func joinFloat64(values []float64) string {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = fmt.Sprintf("%f", v)
	}
	return strings.Join(strs, ",")
}

// GetItems retrieves all items from a collection with optional filtering
func (c *Client) GetItems(ctx context.Context, collectionID string, params *ItemsParams) (*stac.ItemsList, error) {
	if collectionID == "" {
		return nil, fmt.Errorf("collection ID cannot be empty")
	}

	var query url.Values
	if params != nil {
		query = params.buildItemsQuery()
	}

	u := c.buildURL(path.Join("/collections", collectionID, "items"), query)

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	return c.handleItemsRequest(req)
}

// GetItem retrieves a specific item from a collection by ID
func (c *Client) GetItem(ctx context.Context, collectionID, itemID string) (*stac.Item, error) {
	if collectionID == "" {
		return nil, fmt.Errorf("collection ID cannot be empty")
	}
	if itemID == "" {
		return nil, fmt.Errorf("item ID cannot be empty")
	}

	u := c.buildURL(path.Join("/collections", collectionID, "items", itemID), nil)

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("item not found: %s", itemID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var item stac.Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &item, nil
}

// GetNextItems retrieves the next page of items using the `next` rel link
func (c *Client) GetNextItems(ctx context.Context, fc *stac.ItemsList) (*stac.ItemsList, error) {
	for _, l := range fc.Links {
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

func (c *Client) handleItemsRequest(req *http.Request) (*stac.ItemsList, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("item not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var items stac.ItemsList
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &items, nil
}
