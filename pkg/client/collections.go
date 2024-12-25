package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	stac "github.com/planetlabs/go-stac"
)

// Client represents a STAC API client
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new STAC API client
func NewClient(baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Ensure scheme is set
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return &Client{
		BaseURL: u,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}, nil
}

// buildURL constructs a URL for a given path and query parameters
func (c *Client) buildURL(pathname string, query url.Values) *url.URL {
	c.BaseURL.JoinPath(pathname).String()
	u := *c.BaseURL // Create a copy of the base URL
	u.Path = path.Join(u.Path, pathname)
	if query != nil {
		u.RawQuery = query.Encode()
	}
	return &u
}

// GetCollections retrieves all collections from the STAC API
func (c *Client) GetCollections(ctx context.Context) (*stac.CollectionsList, error) {
	u := c.buildURL("/collections", nil)

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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

	var collectionsResp stac.CollectionsList
	if err := json.NewDecoder(resp.Body).Decode(&collectionsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &collectionsResp, nil
}

// GetCollection retrieves a specific collection by ID
func (c *Client) GetCollection(ctx context.Context, collectionID string) (*stac.Collection, error) {
	if collectionID == "" {
		return nil, fmt.Errorf("collection ID cannot be empty")
	}

	u := c.buildURL(path.Join("/collections", collectionID), nil)

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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

	var collection stac.Collection
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

// buildSearchQuery converts SearchCollectionsParams to url.Values
func (p SearchCollectionsParams) buildSearchQuery() url.Values {
	values := url.Values{}

	if p.Limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", p.Limit))
	}
	if p.Query != "" {
		values.Set("query", p.Query)
	}
	for _, field := range p.Fields {
		values.Add("fields", field)
	}

	return values
}

// SearchCollections searches for collections based on provided parameters
func (c *Client) SearchCollections(ctx context.Context, params SearchCollectionsParams) (*stac.CollectionsList, error) {
	u := c.buildURL("/collections", params.buildSearchQuery())

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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

	var collectionsResp stac.CollectionsList
	if err := json.NewDecoder(resp.Body).Decode(&collectionsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &collectionsResp, nil
}
