package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/robert-malhotra/go-stac-client/pkg/stac"
)

// GetCatalog fetches the root catalog document from the STAC API.
// This is typically the entry point for exploring a STAC API, containing
// links to collections, search endpoints, and conformance information.
func (c *Client) GetCatalog(ctx context.Context) (*stac.Catalog, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, c.baseURL.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, c.baseURL)
	}

	var cat stac.Catalog
	err = json.NewDecoder(resp.Body).Decode(&cat)
	return &cat, err
}

// GetConformance fetches the conformance classes supported by the STAC API.
// This is a convenience method that fetches the catalog and returns the conformsTo field.
// For more detailed conformance information, use GetCatalog directly.
func (c *Client) GetConformance(ctx context.Context) ([]string, error) {
	cat, err := c.GetCatalog(ctx)
	if err != nil {
		return nil, err
	}
	return cat.ConformsTo, nil
}

// SupportsConformance checks if the STAC API supports a specific conformance class.
// Use the stac.Conformance* constants for common conformance classes.
func (c *Client) SupportsConformance(ctx context.Context, conformanceClass string) (bool, error) {
	cat, err := c.GetCatalog(ctx)
	if err != nil {
		return false, err
	}
	return cat.HasConformance(conformanceClass), nil
}
