package main

import (
	"encoding/json"

	stac "github.com/planetlabs/go-stac"
)

type itemSummary struct {
	ID         string          `json:"id"`
	Geometry   json.RawMessage `json:"geometry"`
	properties *stac.Properties   `json:"properties"`
	Links      []*stac.Link    `json:"links"`
}

type collectionSummary struct {
	Id          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Extent      *stac.Extent   `json:"extent"`
	Links       []*stac.Link   `json:"links"`
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
