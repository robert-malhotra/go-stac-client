package stacclient_test

import (
	"context"
	"os"
	"testing"
	"time"

	stacclient "github.com/example/go-stac-client/client"
)

func requireLiveEndpoint(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping live STAC test in short mode")
	}
	if os.Getenv("STAC_LIVE_TEST") == "" {
		t.Skip("set STAC_LIVE_TEST=1 to enable live STAC endpoint tests")
	}
	if endpoint := os.Getenv("STAC_LIVE_URL"); endpoint != "" {
		return endpoint
	}
	return "https://earth-search.aws.element84.com/v1"
}

func TestLiveSearchAgainstEndpoint(t *testing.T) {
	endpoint := requireLiveEndpoint(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := stacclient.New(
		stacclient.WithBaseURL(endpoint),
	)
	if err != nil {
		t.Fatalf("New client: %v", err)
	}

	params := stacclient.SearchParams{
		Collections: []string{"sentinel-s2-l2a-cogs"},
		BBox:        []float64{-123.0, 37.0, -121.0, 38.0},
		Datetime:    "2023-01-01T00:00:00Z/2023-02-01T00:00:00Z",
		Limit:       3,
	}

	page, err := client.Search().GetPage(ctx, params)
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if len(page.Items) == 0 {
		t.Fatalf("expected items in live search page, got 0")
	}

	seq := client.Search().Query(ctx, params)
	var count int
	for item, err := range seq {
		if err != nil {
			t.Fatalf("iterator error: %v", err)
		}
		if item == nil {
			continue
		}
		count++
		if count >= 3 {
			break
		}
	}
	if count == 0 {
		t.Fatalf("expected iterator to yield at least one item")
	}
}
