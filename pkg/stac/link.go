package stac

import "encoding/json"

// Link represents a STAC Link with support for additional fields.
type Link struct {
	Href  string `json:"href"`
	Rel   string `json:"rel"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`

	// AdditionalFields holds foreign members (e.g., "method", "body" for POST links).
	AdditionalFields map[string]any `json:"-"`
}

var knownLinkFields = map[string]bool{
	"href": true, "rel": true, "type": true, "title": true,
}

// UnmarshalJSON implements custom unmarshaling to capture foreign members.
func (link *Link) UnmarshalJSON(data []byte) error {
	type linkAlias Link
	var aux linkAlias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*link = Link(aux)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	link.AdditionalFields = make(map[string]any)
	for key, val := range raw {
		if !knownLinkFields[key] {
			var decoded any
			if err := json.Unmarshal(val, &decoded); err != nil {
				continue
			}
			link.AdditionalFields[key] = decoded
		}
	}

	return nil
}

// MarshalJSON implements custom marshaling to include foreign members.
func (link Link) MarshalJSON() ([]byte, error) {
	type linkAlias Link
	aux := linkAlias(link)

	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	if len(link.AdditionalFields) == 0 {
		return data, nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	for key, val := range link.AdditionalFields {
		encoded, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		obj[key] = encoded
	}

	return json.Marshal(obj)
}
