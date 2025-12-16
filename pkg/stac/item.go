package stac

import (
	"encoding/json"
	"fmt"
)

// ItemType is the GeoJSON type for STAC Items (always "Feature").
const ItemType = "Feature"

// Item represents a STAC Item (GeoJSON Feature) with support for foreign members.
// The Type field is implicit and always "Feature" per the GeoJSON/STAC specification.
type Item struct {
	Version    string            `json:"stac_version"`
	Extensions []string          `json:"stac_extensions,omitempty"`
	ID         string            `json:"id"`
	Geometry   any               `json:"geometry"`
	Bbox       []float64         `json:"bbox,omitempty"`
	Properties map[string]any    `json:"properties"`
	Links      []*Link           `json:"links"`
	Assets     map[string]*Asset `json:"assets"`
	Collection string            `json:"collection,omitempty"`

	// AdditionalFields holds foreign members not defined in the STAC spec.
	AdditionalFields map[string]any `json:"-"`
}

var knownItemFields = map[string]bool{
	"type": true, "stac_version": true, "stac_extensions": true,
	"id": true, "geometry": true, "bbox": true, "properties": true,
	"links": true, "assets": true, "collection": true,
}

// UnmarshalJSON implements custom unmarshaling to capture foreign members.
func (item *Item) UnmarshalJSON(data []byte) error {
	type itemAlias Item
	var aux itemAlias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*item = Item(aux)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Validate type field if present
	if typeVal, ok := raw["type"]; ok {
		var t string
		if err := json.Unmarshal(typeVal, &t); err == nil && t != "" && t != ItemType {
			return fmt.Errorf("invalid item type: expected %q, got %q", ItemType, t)
		}
	}

	item.AdditionalFields = make(map[string]any)
	for key, val := range raw {
		if !knownItemFields[key] {
			var decoded any
			if err := json.Unmarshal(val, &decoded); err != nil {
				continue
			}
			item.AdditionalFields[key] = decoded
		}
	}

	return nil
}

// MarshalJSON implements custom marshaling to include foreign members.
// The type field is always set to "Feature" per the GeoJSON/STAC specification.
func (item Item) MarshalJSON() ([]byte, error) {
	type itemAlias Item
	aux := itemAlias(item)

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	// Always include type field
	typeJSON, _ := json.Marshal(ItemType)
	obj["type"] = typeJSON

	// Add foreign members
	for key, val := range item.AdditionalFields {
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		obj[key] = encoded
	}

	return json.Marshal(obj)
}

// GetLink returns the first link with the specified rel type, or nil if not found.
func (item *Item) GetLink(rel string) *Link {
	for _, link := range item.Links {
		if link.Rel == rel {
			return link
		}
	}
	return nil
}

// GetLinks returns all links with the specified rel type.
func (item *Item) GetLinks(rel string) []*Link {
	var result []*Link
	for _, link := range item.Links {
		if link.Rel == rel {
			result = append(result, link)
		}
	}
	return result
}
