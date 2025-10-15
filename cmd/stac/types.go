package main

import (
	"encoding/json"
	"fmt"

	stac "github.com/planetlabs/go-stac"
)

type itemSummary struct {
	ID         string          `json:"id"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties map[string]any  `json:"properties"`
	Links      []*stac.Link    `json:"links"`
}

func newItemSummary(item *stac.Item) (*itemSummary, error) {
	geometry, err := json.Marshal(item.Geometry)
	if err != nil {
		return nil, fmt.Errorf("marshal item geometry: %w", err)
	}

	var properties map[string]any
	if item.Properties != nil {
		properties = make(map[string]any, len(item.Properties))
		for k, v := range item.Properties {
			properties[k] = v
		}
	}

	var links []*stac.Link
	if len(item.Links) > 0 {
		links = make([]*stac.Link, len(item.Links))
		copy(links, item.Links)
	}

	return &itemSummary{
		ID:         item.Id,
		Geometry:   json.RawMessage(geometry),
		Properties: properties,
		Links:      links,
	}, nil
}

type collectionSummary struct {
	Id          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Extent      *stac.Extent `json:"extent"`
	Links       []*stac.Link `json:"links"`
}

func newCollectionSummary(collection *stac.Collection) *collectionSummary {
	return &collectionSummary{
		Id:          collection.Id,
		Title:       collection.Title,
		Description: collection.Description,
		Extent:      collection.Extent,
		Links:       collection.Links,
	}
}
