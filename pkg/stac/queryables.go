package stac

import "encoding/json"

// Queryables represents the queryable properties for a STAC API collection.
// This follows the OGC API - Features - Part 3: Filtering specification.
type Queryables struct {
	Schema      string                      `json:"$schema,omitempty"`
	ID          string                      `json:"$id,omitempty"`
	Type        string                      `json:"type,omitempty"`
	Title       string                      `json:"title,omitempty"`
	Description string                      `json:"description,omitempty"`
	Properties  map[string]*QueryableField `json:"properties,omitempty"`

	// AdditionalFields holds foreign members not defined in the spec.
	AdditionalFields map[string]any `json:"-"`
}

// QueryableField represents a single queryable property with its JSON Schema definition.
type QueryableField struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type,omitempty"`        // "string", "number", "integer", "boolean", "array", "object"
	Format      string   `json:"format,omitempty"`      // e.g., "date-time", "uri"
	Enum        []any    `json:"enum,omitempty"`        // Allowed values
	Minimum     *float64 `json:"minimum,omitempty"`     // For numeric types
	Maximum     *float64 `json:"maximum,omitempty"`     // For numeric types
	MinItems    *int     `json:"minItems,omitempty"`    // For array types
	MaxItems    *int     `json:"maxItems,omitempty"`    // For array types
	Pattern     string   `json:"pattern,omitempty"`     // Regex pattern for strings
	Items       *Items   `json:"items,omitempty"`       // For array types
	Ref         string   `json:"$ref,omitempty"`        // JSON Schema reference
	OneOf       []any    `json:"oneOf,omitempty"`       // Union types
	AnyOf       []any    `json:"anyOf,omitempty"`       // Union types

	// AdditionalFields holds foreign members.
	AdditionalFields map[string]any `json:"-"`
}

// Items represents the items schema for array types.
type Items struct {
	Type string `json:"type,omitempty"`
}

var knownQueryablesFields = map[string]bool{
	"$schema": true, "$id": true, "type": true, "title": true,
	"description": true, "properties": true,
}

var knownQueryableFieldFields = map[string]bool{
	"title": true, "description": true, "type": true, "format": true,
	"enum": true, "minimum": true, "maximum": true, "minItems": true,
	"maxItems": true, "pattern": true, "items": true, "$ref": true,
	"oneOf": true, "anyOf": true,
}

// UnmarshalJSON implements custom unmarshaling to capture foreign members.
func (q *Queryables) UnmarshalJSON(data []byte) error {
	type queryablesAlias Queryables
	var aux queryablesAlias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*q = Queryables(aux)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	q.AdditionalFields = make(map[string]any)
	for key, val := range raw {
		if !knownQueryablesFields[key] {
			var decoded any
			if err := json.Unmarshal(val, &decoded); err != nil {
				continue
			}
			q.AdditionalFields[key] = decoded
		}
	}

	return nil
}

// MarshalJSON implements custom marshaling to include foreign members.
func (q Queryables) MarshalJSON() ([]byte, error) {
	type queryablesAlias Queryables
	aux := queryablesAlias(q)

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	if len(q.AdditionalFields) == 0 {
		return data, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	for key, val := range q.AdditionalFields {
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		obj[key] = encoded
	}

	return json.Marshal(obj)
}

// UnmarshalJSON implements custom unmarshaling for QueryableField.
func (qf *QueryableField) UnmarshalJSON(data []byte) error {
	type fieldAlias QueryableField
	var aux fieldAlias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*qf = QueryableField(aux)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	qf.AdditionalFields = make(map[string]any)
	for key, val := range raw {
		if !knownQueryableFieldFields[key] {
			var decoded any
			if err := json.Unmarshal(val, &decoded); err != nil {
				continue
			}
			qf.AdditionalFields[key] = decoded
		}
	}

	return nil
}

// MarshalJSON implements custom marshaling for QueryableField.
func (qf QueryableField) MarshalJSON() ([]byte, error) {
	type fieldAlias QueryableField
	aux := fieldAlias(qf)

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	if len(qf.AdditionalFields) == 0 {
		return data, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	for key, val := range qf.AdditionalFields {
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		obj[key] = encoded
	}

	return json.Marshal(obj)
}

// DisplayName returns a user-friendly name for the field.
func (qf *QueryableField) DisplayName(key string) string {
	if qf.Title != "" {
		return qf.Title
	}
	return key
}

// TypeDescription returns a human-readable type description.
func (qf *QueryableField) TypeDescription() string {
	if qf.Type == "" {
		return "any"
	}
	desc := qf.Type
	if qf.Format != "" {
		desc += " (" + qf.Format + ")"
	}
	return desc
}
