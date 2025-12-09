package stac

// Extent represents the spatial and temporal extent of a STAC Collection.
type Extent struct {
	Spatial  *SpatialExtent  `json:"spatial,omitempty"`
	Temporal *TemporalExtent `json:"temporal,omitempty"`
}

// SpatialExtent represents the spatial extent of a STAC Collection.
type SpatialExtent struct {
	Bbox [][]float64 `json:"bbox"`
}

// TemporalExtent represents the temporal extent of a STAC Collection.
type TemporalExtent struct {
	Interval [][]any `json:"interval"`
}
