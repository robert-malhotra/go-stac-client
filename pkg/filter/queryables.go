// pkg/filter/queryables.go

package filter

import (
	"encoding/json"
	"errors"
)

// Queryables represents the structure of the queryables JSON Schema
type Queryables struct {
	Schema          string                 `json:"$schema"`
	ID              string                 `json:"$id"`
	Type            string                 `json:"type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Properties      map[string]PropertyRef `json:"properties"`
	AdditionalProps bool                   `json:"additionalProperties"`
}

// PropertyRef represents a single queryable property reference
type PropertyRef struct {
	Description string   `json:"description"`
	Ref         string   `json:"$ref,omitempty"`
	Type        string   `json:"type,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	// Additional fields can be added as needed
}

// ParseQueryables parses JSON data into a Queryables struct
func ParseQueryables(data []byte) (*Queryables, error) {
	var q Queryables
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, err
	}
	return &q, nil
}

// SerializeQueryables serializes a Queryables struct into JSON
func SerializeQueryables(q *Queryables) ([]byte, error) {
	return json.MarshalIndent(q, "", "  ")
}

// ValidateQueryables validates the Queryables struct
func ValidateQueryables(q *Queryables) error {
	if q.Type != "object" {
		return errors.New("queryables must be of type 'object'")
	}
	if q.Properties == nil {
		return errors.New("queryables must have 'properties'")
	}
	// Additional validations can be added here
	return nil
}
