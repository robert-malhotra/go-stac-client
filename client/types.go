package stacclient

import (
	"encoding/json"
	"fmt"
	"net/url"

	ogcfilter "github.com/planetlabs/go-ogc/filter"
)

// CollectionListOption configures a List call.
type CollectionListOption func(*collectionListOptions)

type collectionListOptions struct {
	query          url.Values
	requestOptions []RequestOption
}

func newCollectionListOptions(opts ...CollectionListOption) collectionListOptions {
	cfg := collectionListOptions{query: make(url.Values)}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

// WithCollectionQuery adds a query parameter for collection listing.
func WithCollectionQuery(key, value string) CollectionListOption {
	return func(o *collectionListOptions) {
		if key == "" {
			return
		}
		o.query.Add(key, value)
	}
}

// WithCollectionRequestOption wraps a RequestOption for the list call.
func WithCollectionRequestOption(opt RequestOption) CollectionListOption {
	return func(o *collectionListOptions) {
		if opt != nil {
			o.requestOptions = append(o.requestOptions, opt)
		}
	}
}

// ItemListOption configures item listing requests.
type ItemListOption func(*itemListOptions)

type itemListOptions struct {
	limit          *int
	nextToken      string
	query          url.Values
	requestOptions []RequestOption
}

func newItemListOptions(opts ...ItemListOption) itemListOptions {
	cfg := itemListOptions{query: make(url.Values)}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.limit != nil {
		cfg.query.Set("limit", fmt.Sprint(*cfg.limit))
	}
	if cfg.nextToken != "" {
		cfg.query.Set("token", cfg.nextToken)
	}
	return cfg
}

// WithItemLimit sets the page limit for item listing.
func WithItemLimit(limit int) ItemListOption {
	return func(o *itemListOptions) {
		if limit <= 0 {
			return
		}
		o.limit = &limit
	}
}

// WithItemPageToken sets the pagination token for item listing.
func WithItemPageToken(token string) ItemListOption {
	return func(o *itemListOptions) {
		o.nextToken = token
	}
}

// WithItemQueryParam registers an arbitrary query parameter.
func WithItemQueryParam(key, value string) ItemListOption {
	return func(o *itemListOptions) {
		if key == "" {
			return
		}
		o.query.Add(key, value)
	}
}

// WithItemRequestOption appends a RequestOption for the request.
func WithItemRequestOption(opt RequestOption) ItemListOption {
	return func(o *itemListOptions) {
		if opt != nil {
			o.requestOptions = append(o.requestOptions, opt)
		}
	}
}

// Fields controls property selection in STAC search results.
type Fields struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// SortDirection enumerates search sort orders.
type SortDirection string

const (
	// SortAscending orders results ascending.
	SortAscending SortDirection = "asc"
	// SortDescending orders results descending.
	SortDescending SortDirection = "desc"
)

// SortField describes a search sort clause.
type SortField struct {
	Field     string        `json:"field"`
	Direction SortDirection `json:"direction"`
}

// SearchParams represents the POST body for /search requests.
type SearchParams struct {
	Collections []string                    `json:"collections,omitempty"`
	IDs         []string                    `json:"ids,omitempty"`
	BBox        []float64                   `json:"bbox,omitempty"`
	Datetime    string                      `json:"datetime,omitempty"`
	Intersects  json.RawMessage             `json:"intersects,omitempty"`
	Query       map[string]any              `json:"query,omitempty"`
	SortBy      []SortField                 `json:"sortby,omitempty"`
	Filter      ogcfilter.BooleanExpression `json:"-"`
	FilterText  string                      `json:"-"`
	FilterLang  string                      `json:"filter-lang,omitempty"`
	Fields      *Fields                     `json:"fields,omitempty"`
	Limit       int                         `json:"limit,omitempty"`
	NextToken   string                      `json:"token,omitempty"`
	Additional  map[string]any              `json:"-"`
}

func (p SearchParams) body() (map[string]any, error) {
	body := map[string]any{}
	if len(p.Collections) > 0 {
		body["collections"] = p.Collections
	}
	if len(p.IDs) > 0 {
		body["ids"] = p.IDs
	}
	if len(p.BBox) > 0 {
		body["bbox"] = p.BBox
	}
	if p.Datetime != "" {
		body["datetime"] = p.Datetime
	}
	if len(p.Query) > 0 {
		body["query"] = p.Query
	}
	if len(p.SortBy) > 0 {
		body["sortby"] = p.SortBy
	}
	if p.Fields != nil {
		body["fields"] = p.Fields
	}
	if p.Limit > 0 {
		body["limit"] = p.Limit
	}
	if p.NextToken != "" {
		body["token"] = p.NextToken
	}
	if len(p.Intersects) > 0 {
		body["intersects"] = p.Intersects
	}
	if p.Filter != nil {
		data, err := json.Marshal(p.Filter)
		if err != nil {
			return nil, err
		}
		var decoded any
		if err := json.Unmarshal(data, &decoded); err != nil {
			return nil, err
		}
		body["filter"] = decoded
		if p.FilterLang == "" {
			body["filter-lang"] = "cql2-json"
		}
	}
	if p.FilterText != "" {
		body["filter"] = p.FilterText
		if p.FilterLang == "" {
			body["filter-lang"] = "cql2-text"
		}
	}
	if p.FilterLang != "" {
		body["filter-lang"] = p.FilterLang
	}
	for key, value := range p.Additional {
		body[key] = value
	}
	return body, nil
}

// Clone returns a copy of the SearchParams with shallow-copied slices/maps.
func (p SearchParams) Clone() SearchParams {
	cp := p
	if len(p.Collections) > 0 {
		cp.Collections = append([]string{}, p.Collections...)
	}
	if len(p.IDs) > 0 {
		cp.IDs = append([]string{}, p.IDs...)
	}
	if len(p.BBox) > 0 {
		cp.BBox = append([]float64{}, p.BBox...)
	}
	if len(p.SortBy) > 0 {
		cp.SortBy = append([]SortField{}, p.SortBy...)
	}
	if p.Fields != nil {
		fields := *p.Fields
		if len(fields.Include) > 0 {
			fields.Include = append([]string{}, fields.Include...)
		}
		if len(fields.Exclude) > 0 {
			fields.Exclude = append([]string{}, fields.Exclude...)
		}
		cp.Fields = &fields
	}
	if p.Query != nil {
		cp.Query = make(map[string]any, len(p.Query))
		for k, v := range p.Query {
			cp.Query[k] = v
		}
	}
	if p.Additional != nil {
		cp.Additional = make(map[string]any, len(p.Additional))
		for k, v := range p.Additional {
			cp.Additional[k] = v
		}
	}
	if len(p.Intersects) > 0 {
		cp.Intersects = append([]byte{}, p.Intersects...)
	}
	return cp
}
