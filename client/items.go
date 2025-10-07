package stacclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	stac "github.com/planetlabs/go-stac"
	"iter"
)

// ItemService provides item listing and retrieval operations.
type ItemService struct {
	client *Client
}

// Get streams items from a collection as an iterator sequence.
func (s *ItemService) Get(ctx context.Context, collectionID string, opts ...ItemListOption) iter.Seq2[*stac.Item, error] {
	if collectionID == "" {
		return func(yield func(*stac.Item, error) bool) {
			yield(nil, fmt.Errorf("collection id is required"))
		}
	}

	cfg := newItemListOptions(opts...)
	baseQuery := cloneValues(cfg.query)
	if baseQuery == nil {
		baseQuery = make(url.Values)
	}
	baseQuery.Del("token")
	requestOpts := append([]RequestOption{}, cfg.requestOptions...)
	limit := cfg.limit
	initialToken := cfg.nextToken

	return func(yield func(*stac.Item, error) bool) {
		token := initialToken
		for {
			query := cloneValues(baseQuery)
			if query == nil {
				query = make(url.Values)
			}
			if limit != nil {
				query.Set("limit", fmt.Sprint(*limit))
			}
			if token != "" {
				query.Set("token", token)
			}
			page, err := s.fetchPage(ctx, collectionID, query, requestOpts)
			if err != nil {
				yield(nil, err)
				return
			}
			for _, item := range page.Items {
				if item == nil {
					continue
				}
				if !yield(item, nil) {
					return
				}
			}
			nextToken := page.NextToken()
			if nextToken == "" {
				return
			}
			token = nextToken
		}
	}
}

// GetPage fetches a single page of items.
func (s *ItemService) GetPage(ctx context.Context, collectionID string, opts ...ItemListOption) (*ItemCollection, error) {
	if collectionID == "" {
		return nil, fmt.Errorf("collection id is required")
	}
	cfg := newItemListOptions(opts...)
	query := cloneValues(cfg.query)
	requestOpts := append([]RequestOption{}, cfg.requestOptions...)
	return s.fetchPage(ctx, collectionID, query, requestOpts)
}

// GetOne retrieves a single item from a collection by ID.
func (s *ItemService) GetOne(ctx context.Context, collectionID, itemID string, opts ...RequestOption) (*stac.Item, error) {
	if collectionID == "" {
		return nil, fmt.Errorf("collection id is required")
	}
	if itemID == "" {
		return nil, fmt.Errorf("item id is required")
	}
	endpoint := fmt.Sprintf("/collections/%s/items/%s", url.PathEscape(collectionID), url.PathEscape(itemID))
	var item stac.Item
	if err := s.client.doJSON(ctx, http.MethodGet, endpoint, nil, nil, &item, opts); err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ItemService) fetchPage(ctx context.Context, collectionID string, query url.Values, opts []RequestOption) (*ItemCollection, error) {
	endpoint := fmt.Sprintf("/collections/%s/items", url.PathEscape(collectionID))
	var page ItemCollection
	if err := s.client.doJSON(ctx, http.MethodGet, endpoint, query, nil, &page, opts); err != nil {
		return nil, err
	}
	return &page, nil
}
