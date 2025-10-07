package auth

import "net/http"

// APIKeyTransport injects an API key header into outgoing requests.
type APIKeyTransport struct {
	Key    string
	Header string
	Base   http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (t *APIKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	header := t.Header
	if header == "" {
		header = "Authorization"
	}
	if t.Key != "" {
		clone.Header.Set(header, t.Key)
	}
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}

// BearerTokenTransport injects a bearer token.
type BearerTokenTransport struct {
	Token string
	Base  http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (t *BearerTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	if t.Token != "" {
		clone.Header.Set("Authorization", "Bearer "+t.Token)
	}
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(clone)
}
