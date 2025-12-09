package stac

import "encoding/json"

// Item represents a STAC Item (GeoJSON Feature) with support for foreign members.
type Item struct {
	Type       string            `json:"type,omitempty"`
	Version    string            `json:"stac_version"`
	Extensions []string          `json:"stac_extensions,omitempty"`
	Id         string            `json:"id"`
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
func (item Item) MarshalJSON() ([]byte, error) {
	type itemAlias Item
	aux := itemAlias(item)

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	if len(item.AdditionalFields) == 0 {
		return data, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	for key, val := range item.AdditionalFields {
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		obj[key] = encoded
	}

	return json.Marshal(obj)
}
