package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"

	stac "github.com/planetlabs/go-stac"
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
	return iteratePages[stac.Collection](ctx, c, "collections",
		func(r io.Reader) ([]*stac.Collection, []*stac.Link, error) {
			var page struct {
				Collections []*stac.Collection `json:"collections"`
				Links       []*stac.Link       `json:"links"`
			}
			err := json.NewDecoder(r).Decode(&page)
			return page.Collections, page.Links, err
		})
}
