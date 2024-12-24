// File: pkg/stac/items.go
package stac

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Item represents a STAC item
type Item struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Geometry   map[string]interface{} `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
	Links      []Link                 `json:"links"`
	Assets     map[string]Asset       `json:"assets"`
}

// Asset represents a STAC item asset
type Asset struct {
	Href  string   `json:"href"`
	Type  string   `json:"type,omitempty"`
	Title string   `json:"title,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// ItemsResponse represents the response from an items endpoint
type ItemsResponse struct {
	Type     string `json:"type"`
	Features []Item `json:"features"`
	Links    []Link `json:"links"`
}

// SearchItemsParams represents the parameters for searching items
type SearchItemsParams struct {
	Collections []string               `json:"collections,omitempty"`
	BBox        []float64              `json:"bbox,omitempty"`
	Datetime    string                 `json:"datetime,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Query       map[string]interface{} `json:"query,omitempty"`
	Filter      map[string]interface{} `json:"filter,omitempty"` // CQL2-JSON filter
}

// GetCollectionItems retrieves items from a specific collection
func (c *Client) GetCollectionItems(ctx context.Context, collectionID string) (*ItemsResponse, error) {
	endpoint := fmt.Sprintf("%s/collections/%s/items", c.BaseURL, url.PathEscape(collectionID))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &itemsResp, nil
}

// GetItem retrieves a specific item from a collection
func (c *Client) GetItem(ctx context.Context, collectionID, itemID string) (*Item, error) {
	endpoint := fmt.Sprintf("%s/collections/%s/items/%s", c.BaseURL, url.PathEscape(collectionID), url.PathEscape(itemID))

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var item Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &item, nil
}

// SearchItems searches for items across collections with parameters
func (c *Client) SearchItems(ctx context.Context, params SearchItemsParams) (*ItemsResponse, error) {
	endpoint := fmt.Sprintf("%s/search", c.BaseURL)

	// Create request body for POST
	requestBody := make(map[string]interface{})

	if len(params.Collections) > 0 {
		requestBody["collections"] = params.Collections
	}
	if len(params.BBox) > 0 {
		requestBody["bbox"] = params.BBox
	}
	if params.Datetime != "" {
		requestBody["datetime"] = params.Datetime
	}
	if params.Limit > 0 {
		requestBody["limit"] = params.Limit
	}
	if params.Filter != nil {
		requestBody["filter"] = params.Filter
	}
	if params.Query != nil {
		requestBody["query"] = params.Query
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &itemsResp, nil
}

func formatBBox(bbox []float64) string {
	var result string
	for i, v := range bbox {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%f", v)
	}
	return result
}
