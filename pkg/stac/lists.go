package stac

// ItemsList represents a GeoJSON FeatureCollection of STAC Items.
type ItemsList struct {
	Type     string  `json:"type"`
	Features []*Item `json:"features"`
	Links    []*Link `json:"links,omitempty"`
}

// CollectionsList represents a list of STAC Collections.
type CollectionsList struct {
	Collections []*Collection `json:"collections"`
	Links       []*Link       `json:"links,omitempty"`
}
