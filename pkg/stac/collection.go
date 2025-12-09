package stac

import "encoding/json"

// Collection represents a STAC Collection with support for foreign members.
type Collection struct {
	Type        string            `json:"type,omitempty"`
	Version     string            `json:"stac_version"`
	Extensions  []string          `json:"stac_extensions,omitempty"`
	Id          string            `json:"id"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords,omitempty"`
	License     string            `json:"license"`
	Providers   []*Provider       `json:"providers,omitempty"`
	Extent      *Extent           `json:"extent"`
	Summaries   map[string]any    `json:"summaries,omitempty"`
	Links       []*Link           `json:"links"`
	Assets      map[string]*Asset `json:"assets,omitempty"`

	// AdditionalFields holds foreign members not defined in the STAC spec.
	AdditionalFields map[string]any `json:"-"`
}

var knownCollectionFields = map[string]bool{
	"type": true, "stac_version": true, "stac_extensions": true,
	"id": true, "title": true, "description": true, "keywords": true,
	"license": true, "providers": true, "extent": true, "summaries": true,
	"links": true, "assets": true,
}

// UnmarshalJSON implements custom unmarshaling to capture foreign members.
func (col *Collection) UnmarshalJSON(data []byte) error {
	type collectionAlias Collection
	var aux collectionAlias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*col = Collection(aux)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	col.AdditionalFields = make(map[string]any)
	for key, val := range raw {
		if !knownCollectionFields[key] {
			var decoded any
			if err := json.Unmarshal(val, &decoded); err != nil {
				continue
			}
			col.AdditionalFields[key] = decoded
		}
	}

	return nil
}

// MarshalJSON implements custom marshaling to include foreign members.
func (col Collection) MarshalJSON() ([]byte, error) {
	type collectionAlias Collection
	aux := collectionAlias(col)

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	if len(col.AdditionalFields) == 0 {
		return data, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	for key, val := range col.AdditionalFields {
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		obj[key] = encoded
	}

	return json.Marshal(obj)
}
