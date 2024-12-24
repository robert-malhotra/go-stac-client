// File: pkg/stac/stac.go
package stac

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a STAC API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new STAC API client
func NewClient(baseURL string) *Client {
	// Ensure the base URL has a scheme
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	// Ensure the base URL doesn't end with a trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// Collection represents a STAC collection
type Collection struct {
	ID          string     `json:"id"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description"`
	Keywords    []string   `json:"keywords,omitempty"`
	Version     string     `json:"version,omitempty"`
	License     string     `json:"license"`
	Providers   []Provider `json:"providers,omitempty"`
	Extent      Extent     `json:"extent"`
	Links       []Link     `json:"links"`
}

// Provider represents a STAC provider
type Provider struct {
	Name  string   `json:"name"`
	URL   string   `json:"url,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// Extent represents the spatial and temporal extent of a collection
type Extent struct {
	Spatial  SpatialExtent  `json:"spatial"`
	Temporal TemporalExtent `json:"temporal"`
}

// SpatialExtent represents the spatial bounds of a collection
type SpatialExtent struct {
	BoundingBox [][4]float64 `json:"bbox"`
}

// TemporalExtent represents the temporal bounds of a collection
type TemporalExtent struct {
	Interval [][2]string `json:"interval"`
}

// Link represents a STAC link
type Link struct {
	Href  string `json:"href"`
	Rel   string `json:"rel"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

// CollectionsResponse represents the response from a collections endpoint
type CollectionsResponse struct {
	Collections []Collection `json:"collections"`
	Links       []Link       `json:"links"`
}

// GetCollections retrieves all collections from the STAC API
func (c *Client) GetCollections(ctx context.Context) (*CollectionsResponse, error) {
	endpoint := fmt.Sprintf("%s/collections", c.BaseURL)

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

	var collectionsResp CollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &collectionsResp, nil
}

// GetCollection retrieves a specific collection by ID
func (c *Client) GetCollection(ctx context.Context, collectionID string) (*Collection, error) {
	endpoint := fmt.Sprintf("%s/collections/%s", c.BaseURL, url.PathEscape(collectionID))

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

	var collection Collection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &collection, nil
}

// SearchCollectionsParams represents the parameters for searching collections
type SearchCollectionsParams struct {
	Limit  int      `json:"limit,omitempty"`
	Query  string   `json:"query,omitempty"`
	Fields []string `json:"fields,omitempty"`
}

// SearchCollections searches for collections based on provided parameters
func (c *Client) SearchCollections(ctx context.Context, params SearchCollectionsParams) (*CollectionsResponse, error) {
	endpoint := fmt.Sprintf("%s/collections", c.BaseURL)

	// Build query parameters
	values := url.Values{}
	if params.Limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Query != "" {
		values.Set("query", params.Query)
	}
	if len(params.Fields) > 0 {
		for _, field := range params.Fields {
			values.Add("fields", field)
		}
	}

	if len(values) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, values.Encode())
	}

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

	var collectionsResp CollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &collectionsResp, nil
}
