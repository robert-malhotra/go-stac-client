package client

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/url"

	"github.com/robert-malhotra/go-stac-client/pkg/stac"
)

// GetItem fetches an individual item from a collection.
func (c *Client) GetItem(ctx context.Context, collectionID, itemID string) (*stac.Item, error) {
	if collectionID == "" {
		return nil, fmt.Errorf("collection ID cannot be empty")
	}
	if itemID == "" {
		return nil, fmt.Errorf("item ID cannot be empty")
	}

	u := c.baseURL.JoinPath("collections", collectionID, "items", itemID)

	resp, err := c.doRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var item stac.Item
		if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
			return nil, fmt.Errorf("error decoding response from %s: %w", u, err)
		}
		return &item, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("item not found: %s", itemID)
	default:
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, u)
	}
}

func (c *Client) GetItems(ctx context.Context, collectionID string) iter.Seq2[*stac.Item, error] {
	return c.GetItemsWithDecoder(ctx, collectionID, DefaultItemDecoder())
}

// GetItemsWithDecoder fetches items from a collection using a custom page decoder.
// This is useful for APIs that return non-standard response formats.
func (c *Client) GetItemsWithDecoder(ctx context.Context, collectionID string, decoder PageDecoder[stac.Item]) iter.Seq2[*stac.Item, error] {
	if collectionID == "" {
		return func(y func(*stac.Item, error) bool) {
			y(nil, fmt.Errorf("collection ID cannot be empty"))
		}
	}

	start := fmt.Sprintf("collections/%s/items", url.PathEscape(collectionID))

	return iteratePagesWithDecoder[stac.Item](ctx, c, start, decoder)
}

// GetItemsFromPath fetches items from an arbitrary path using a custom page decoder.
// This is useful for APIs with non-standard endpoint paths (e.g., ICEYE's /catalog/v2/items).
func (c *Client) GetItemsFromPath(ctx context.Context, path string, decoder PageDecoder[stac.Item]) iter.Seq2[*stac.Item, error] {
	return iteratePagesWithDecoder[stac.Item](ctx, c, path, decoder)
}
