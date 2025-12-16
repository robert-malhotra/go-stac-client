package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"iter"

	"github.com/robert-malhotra/go-stac-client/pkg/stac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collect pulls every collection from the iterator until it terminates.
func collectCollections(seq iter.Seq2[*stac.Collection, error]) ([]*stac.Collection, error) {
	var (
		out    []*stac.Collection
		outErr error
	)
	seq(func(col *stac.Collection, err error) bool {
		if err != nil {
			outErr = err
			return false
		}
		out = append(out, col)
		return true
	})
	return out, outErr
}

// newMockServer returns a server for collections endpoints.
func newCollectionMockServer(t *testing.T) *httptest.Server {
	list := fixture(t, "collections.json")
	single := fixture(t, "single_collection.json")
	bad := []byte(`{"broken":`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// trim leading slash
		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		// /collections or /collections/{id}
		if len(parts) >= 1 && parts[0] == "collections" {
			// list collections
			if len(parts) == 1 {
				// decode error scenario via query param ?bad
				if r.URL.Query().Get("bad") == "1" {
					w.Write(bad)
					return
				}
				w.Write(list)
				return
			}
			// single collection
			if len(parts) == 2 {
				switch parts[1] {
				case "sentinel-2":
					w.Write(single)
				case "error-collection":
					http.Error(w, "boom", http.StatusInternalServerError)
				case "bad-collection":
					w.Write(bad)
				default:
					http.NotFound(w, r)
				}
				return
			}
		}
		http.NotFound(w, r)
	})
	return httptest.NewServer(handler)
}

func TestClient_GetCollection(t *testing.T) {
	mock := newCollectionMockServer(t)
	defer mock.Close()

	cli, err := NewClient(mock.URL)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		col, err := cli.GetCollection(context.Background(), "sentinel-2")
		require.NoError(t, err)
		assert.Equal(t, "sentinel-2", col.ID)
	})

	t.Run("empty ID", func(t *testing.T) {
		_, err := cli.GetCollection(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "collection ID cannot be empty")
	})

	t.Run("not found", func(t *testing.T) {
		_, err := cli.GetCollection(context.Background(), "missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code 404")
	})

	t.Run("server error", func(t *testing.T) {
		_, err := cli.GetCollection(context.Background(), "error-collection")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code 500")
	})

}

func TestClient_GetCollections(t *testing.T) {
	mock := newCollectionMockServer(t)
	defer mock.Close()

	cli, err := NewClient(mock.URL)
	require.NoError(t, err)

	t.Run("list collections", func(t *testing.T) {
		cols, err := collectCollections(cli.GetCollections(context.Background()))
		require.NoError(t, err)
		require.Len(t, cols, 2)
		assert.Equal(t, "sentinel-2", cols[0].ID)
		assert.Equal(t, "landsat-8", cols[1].ID)
	})

	t.Run("early stop", func(t *testing.T) {
		seq := cli.GetCollections(context.Background())
		count := 0
		seq(func(col *stac.Collection, err error) bool {
			require.NoError(t, err)
			count++
			return false
		})
		assert.Equal(t, 1, count)
	})

}

func TestLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in -short mode")
	}

	candidates := []string{
		"https://planetarycomputer.microsoft.com/api/stac/v1",
		"https://earth-search.aws.element84.com/v1",
		"https://cmr.earthdata.nasa.gov/stac/ASF",
	}

	for _, base := range candidates {
		cli, err := NewClient(base)
		if err != nil {
			t.Logf("live test: skipping candidate %s (client init failed: %v)", base, err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		seq := cli.GetCollections(ctx)
		var (
			count   int
			iterErr error
		)

		seq(func(col *stac.Collection, err error) bool {
			if err != nil {
				iterErr = err
				return false
			}
			count++
			t.Logf("%s collection[%d]=%s", base, count, col.ID)
			return count < 5
		})
		cancel()

		if iterErr != nil {
			t.Logf("live test: candidate %s returned error: %v", base, iterErr)
			continue
		}
		if count == 0 {
			t.Logf("live test: candidate %s returned zero collections", base)
			continue
		}

		t.Logf("live test succeeded against %s (count=%d)", base, count)
		return
	}

	t.Skip("skipping live test: no candidate STAC API responded successfully")
}
