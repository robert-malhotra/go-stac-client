package stac

import (
	"encoding/json"
	"fmt"
)

// CollectionType is the STAC type for Collections (always "Collection").
const CollectionType = "Collection"

// Collection represents a STAC Collection with support for foreign members.
// The Type field is implicit and always "Collection" per the STAC specification.
type Collection struct {
	Version     string            `json:"stac_version"`
	Extensions  []string          `json:"stac_extensions,omitempty"`
	ID          string            `json:"id"`
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

	// Validate type field if present
	if typeVal, ok := raw["type"]; ok {
		var t string
		if err := json.Unmarshal(typeVal, &t); err == nil && t != "" && t != CollectionType {
			return fmt.Errorf("invalid collection type: expected %q, got %q", CollectionType, t)
		}
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
// The type field is always set to "Collection" per the STAC specification.
func (col Collection) MarshalJSON() ([]byte, error) {
	type collectionAlias Collection
	aux := collectionAlias(col)

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	// Always include type field
	typeJSON, _ := json.Marshal(CollectionType)
	obj["type"] = typeJSON

	// Add foreign members
	for key, val := range col.AdditionalFields {
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		obj[key] = encoded
	}

	return json.Marshal(obj)
}

// GetLink returns the first link with the specified rel type, or nil if not found.
func (col *Collection) GetLink(rel string) *Link {
	for _, link := range col.Links {
		if link.Rel == rel {
			return link
		}
	}
	return nil
}

// GetLinks returns all links with the specified rel type.
func (col *Collection) GetLinks(rel string) []*Link {
	var result []*Link
	for _, link := range col.Links {
		if link.Rel == rel {
			result = append(result, link)
		}
	}
	return result
}
