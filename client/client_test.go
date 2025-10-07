package stacclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	stacclient "github.com/example/go-stac-client/client"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *stacclient.Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := stacclient.New(
		stacclient.WithBaseURL(server.URL),
		stacclient.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("New client: %v", err)
	}
	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

type itemPayload struct {
	Type    string        `json:"type"`
	Items   []any         `json:"features"`
	Links   []linkPayload `json:"links,omitempty"`
	Context any           `json:"context,omitempty"`
}

type linkPayload struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

func minimalItem(id string) map[string]any {
	return map[string]any{
		"type":         "Feature",
		"stac_version": "1.0.0",
		"id":           id,
		"geometry":     nil,
		"bbox":         nil,
		"properties":   map[string]any{},
		"assets":       map[string]any{},
		"links":        []any{},
	}
}

func TestItemsIterator(t *testing.T) {
	var requests int
	handler := func(w http.ResponseWriter, r *http.Request) {
		requests++
		token := r.URL.Query().Get("token")
		switch token {
		case "":
			nextURL := "http://" + r.Host + r.URL.Path + "?token=abc"
			writeJSON(t, w, itemPayload{
				Type:  "FeatureCollection",
				Items: []any{minimalItem("one")},
				Links: []linkPayload{{Rel: "next", Href: nextURL}},
			})
		case "abc":
			writeJSON(t, w, itemPayload{
				Type:  "FeatureCollection",
				Items: []any{minimalItem("two")},
			})
		default:
			t.Fatalf("unexpected token %q", token)
		}
	}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/demo/items" {
			handler(w, r)
			return
		}
		http.NotFound(w, r)
	})

	seq := client.Items().Get(context.Background(), "demo")
	var ids []string
	for item, err := range seq {
		if err != nil {
			t.Fatalf("iterator error: %v", err)
		}
		ids = append(ids, item.Id)
	}

	if got, want := len(ids), 2; got != want {
		t.Fatalf("expected %d items, got %d", want, got)
	}
	if ids[0] != "one" || ids[1] != "two" {
		t.Fatalf("unexpected ids: %#v", ids)
	}
	if requests != 2 {
		t.Fatalf("expected 2 requests, got %d", requests)
	}
}

func TestSearchIterator(t *testing.T) {
	var tokens []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			tok, _ := payload["token"].(string)
			tokens = append(tokens, tok)
			var items []any
			var links []linkPayload
			switch tok {
			case "":
				items = []any{minimalItem("alpha")}
				links = []linkPayload{{Rel: "next", Href: "http://" + r.Host + "/search?token=next"}}
			case "next":
				items = []any{minimalItem("beta")}
			}
			writeJSON(t, w, itemPayload{Type: "FeatureCollection", Items: items, Links: links})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := stacclient.New(
		stacclient.WithBaseURL(server.URL),
		stacclient.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("New client: %v", err)
	}

	params := stacclient.SearchParams{Limit: 1}
	seq := client.Search().Query(context.Background(), params)
	var ids []string
	for item, err := range seq {
		if err != nil {
			t.Fatalf("iterator error: %v", err)
		}
		ids = append(ids, item.Id)
	}
	if got, want := len(ids), 2; got != want {
		t.Fatalf("expected %d results, got %d", want, got)
	}
	if len(tokens) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(tokens))
	}
	if tokens[0] != "" || tokens[1] != "next" {
		t.Fatalf("unexpected tokens: %#v", tokens)
	}
}
