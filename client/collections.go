package stacclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	stac "github.com/planetlabs/go-stac"
)

// CollectionListResult represents the response for /collections.
type CollectionListResult struct {
	Collections []*stac.Collection `json:"collections"`
	Links       []stac.Link        `json:"links,omitempty"`
}

// CollectionService provides access to STAC collections.
type CollectionService struct {
	client *Client
}

// Get retrieves a single collection by ID.
func (s *CollectionService) Get(ctx context.Context, id string, opts ...RequestOption) (*stac.Collection, error) {
	if id == "" {
		return nil, fmt.Errorf("collection id is required")
	}
	endpoint := fmt.Sprintf("/collections/%s", url.PathEscape(id))
	var collection stac.Collection
	if err := s.client.doJSON(ctx, http.MethodGet, endpoint, nil, nil, &collection, opts); err != nil {
		return nil, err
	}
	return &collection, nil
}

// List retrieves all collections with optional query parameters.
func (s *CollectionService) List(ctx context.Context, opts ...CollectionListOption) (*CollectionListResult, error) {
	cfg := newCollectionListOptions(opts...)
	query := cloneValues(cfg.query)
	requestOpts := append([]RequestOption{}, cfg.requestOptions...)

	var result CollectionListResult
	if err := s.client.doJSON(ctx, http.MethodGet, "/collections", query, nil, &result, requestOpts); err != nil {
		return nil, err
	}
	return &result, nil
}
