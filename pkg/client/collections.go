package client

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"

	"github.com/robert-malhotra/go-stac-client/pkg/stac"
)

// GetCollection fetches a single collection document by ID.
func (c *Client) GetCollection(ctx context.Context, collectionID string) (*stac.Collection, error) {
	if collectionID == "" {
		return nil, fmt.Errorf("collection ID cannot be empty")
	}

	u := c.baseURL.JoinPath("collections", collectionID)

	resp, err := c.doRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, u)
	}

	var col stac.Collection
	err = json.NewDecoder(resp.Body).Decode(&col)
	return &col, err
}

// GetCollections iterates over every collection exposed by the STAC API
// referenced by the client. It transparently follows pagination using the
// client's nextHandler. The returned iterator yields either *stac.Collection
// values or an error. The iteration stops when the consumer returns false or
// when there are no further pages.
func (c *Client) GetCollections(ctx context.Context) iter.Seq2[*stac.Collection, error] {
	return c.GetCollectionsWithDecoder(ctx, DefaultCollectionDecoder())
}

// GetCollectionsWithDecoder fetches collections using a custom page decoder.
// This is useful for APIs that return non-standard response formats.
func (c *Client) GetCollectionsWithDecoder(ctx context.Context, decoder PageDecoder[stac.Collection]) iter.Seq2[*stac.Collection, error] {
	return iteratePagesWithDecoder[stac.Collection](ctx, c, "collections", decoder)
}

// GetQueryables fetches the queryable properties for a collection.
// The endpoint is /collections/{collectionId}/queryables as per OGC API - Features Part 3.
func (c *Client) GetQueryables(ctx context.Context, collectionID string) (*stac.Queryables, error) {
	if collectionID == "" {
		return nil, fmt.Errorf("collection ID cannot be empty")
	}

	u := c.baseURL.JoinPath("collections", collectionID, "queryables")

	resp, err := c.doRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("queryables not available for collection %s", collectionID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, u)
	}

	var q stac.Queryables
	err = json.NewDecoder(resp.Body).Decode(&q)
	return &q, err
}

// GetGlobalQueryables fetches the global queryable properties for the STAC API.
// The endpoint is /queryables as per OGC API - Features Part 3.
func (c *Client) GetGlobalQueryables(ctx context.Context) (*stac.Queryables, error) {
	u := c.baseURL.JoinPath("queryables")

	resp, err := c.doRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("queryables endpoint not available")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, u)
	}

	var q stac.Queryables
	err = json.NewDecoder(resp.Body).Decode(&q)
	return &q, err
}
