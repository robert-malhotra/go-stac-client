package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"

	stac "github.com/planetlabs/go-stac"
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
	if collectionID == "" {
		return func(y func(*stac.Item, error) bool) {
			y(nil, fmt.Errorf("collection ID cannot be empty"))
		}
	}

	start := fmt.Sprintf("collections/%s/items", url.PathEscape(collectionID))

	return iteratePages[stac.Item](ctx, c, start,
		func(r io.Reader) ([]*stac.Item, []*stac.Link, error) {
			var page struct {
				Features []*stac.Item `json:"features"` // STAC 1.0 Items list
				Links    []*stac.Link `json:"links"`
			}
			err := json.NewDecoder(r).Decode(&page)
			return page.Features, page.Links, err
		})
}
