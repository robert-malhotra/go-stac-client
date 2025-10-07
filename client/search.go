package stacclient

import (
	"context"
	"fmt"
	"net/http"

	stac "github.com/planetlabs/go-stac"
	"iter"
)

// SearchService executes STAC search requests.
type SearchService struct {
	client *Client
}

// Query streams search results lazily using STAC pagination tokens.
func (s *SearchService) Query(ctx context.Context, params SearchParams, opts ...RequestOption) iter.Seq2[*stac.Item, error] {
	base := params.Clone()
	return func(yield func(*stac.Item, error) bool) {
		token := base.NextToken
		for {
			current := base.Clone()
			current.NextToken = token
			page, err := s.GetPage(ctx, current, opts...)
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
			next := page.NextToken()
			if next == "" {
				return
			}
			token = next
		}
	}
}

// GetPage performs a single POST /search request returning one page of items.
func (s *SearchService) GetPage(ctx context.Context, params SearchParams, opts ...RequestOption) (*ItemCollection, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	body, err := params.body()
	if err != nil {
		return nil, err
	}
	var page ItemCollection
	if err := s.client.doJSON(ctx, http.MethodPost, "/search", nil, body, &page, opts); err != nil {
		return nil, err
	}
	return &page, nil
}

// Validate ensures the provided search parameters are usable.
func (p SearchParams) Validate() error {
	if len(p.BBox) != 0 && len(p.BBox) != 4 && len(p.BBox) != 6 {
		return fmt.Errorf("bbox must contain 4 or 6 coordinates")
	}
	return nil
}
