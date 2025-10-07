package stacclient

import (
	"net/url"
	"strings"

	stac "github.com/planetlabs/go-stac"
)

// ItemCollection represents a STAC ItemCollection response.
type ItemCollection struct {
	Type    string         `json:"type"`
	Items   []*stac.Item   `json:"features"`
	Links   []*stac.Link   `json:"links,omitempty"`
	Context map[string]any `json:"context,omitempty"`
	Extras  map[string]any `json:"-"`
}

// NextLink returns the rel="next" link if present.
func (c *ItemCollection) NextLink() *stac.Link {
	if c == nil {
		return nil
	}
	for _, link := range c.Links {
		if link == nil {
			continue
		}
		if strings.EqualFold(link.Rel, "next") {
			return link
		}
	}
	return nil
}

// NextToken extracts a token query parameter from the next link, if present.
func (c *ItemCollection) NextToken() string {
	link := c.NextLink()
	if link == nil || link.Href == "" {
		return ""
	}
	u, err := url.Parse(link.Href)
	if err != nil {
		return ""
	}
	return u.Query().Get("token")
}
