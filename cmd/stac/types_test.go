package main

import (
	"encoding/json"
	"testing"

	stac "github.com/planetlabs/go-stac"
	"github.com/stretchr/testify/require"
)

func TestNewItemSummary(t *testing.T) {
	item := &stac.Item{
		Id: "item-123",
		Geometry: map[string]any{
			"type":        "Point",
			"coordinates": []float64{1, 2},
		},
		Properties: map[string]any{
			"prop":   "value",
			"nested": map[string]any{"foo": "bar"},
		},
		Links: []*stac.Link{{Rel: "self", Href: "http://example.com/items/item-123"}},
	}

	summary, err := newItemSummary(item)
	require.NoError(t, err)
	require.Equal(t, "item-123", summary.ID)
	require.NotNil(t, summary.Properties)
	require.Equal(t, item.Properties, summary.Properties)

	var geometry map[string]any
	require.NoError(t, json.Unmarshal(summary.Geometry, &geometry))
	require.Equal(t, "Point", geometry["type"])
	coords, ok := geometry["coordinates"].([]any)
	require.True(t, ok)
	require.Len(t, coords, 2)
	require.Equal(t, 1.0, coords[0])
	require.Equal(t, 2.0, coords[1])

	data, err := json.Marshal(summary)
	require.NoError(t, err)

	var roundTrip map[string]any
	require.NoError(t, json.Unmarshal(data, &roundTrip))
	require.Contains(t, roundTrip, "properties")
	props, ok := roundTrip["properties"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "value", props["prop"])

	summary.Properties["prop"] = "changed"
	require.Equal(t, "value", item.Properties["prop"])
}
