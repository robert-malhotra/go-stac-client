package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDataDir = ""

// loadTestData loads a JSON file from the testdata directory
func loadTestData(t *testing.T, filename string) []byte {
	path := filepath.Join(testDataDir, filename)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read test data file")
	return data
}

type mockServer struct {
	*httptest.Server
	Collections      []byte
	SingleCollection []byte
}

func newMockServer(t *testing.T) *mockServer {
	ms := &mockServer{
		Collections:      loadTestData(t, "collections.json"),
		SingleCollection: loadTestData(t, "single_collection.json"),
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/collections":
			w.Write(ms.Collections)
		case "/collections/sentinel-2":
			w.Write(ms.SingleCollection)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return ms
}

func TestGetCollections(t *testing.T) {
	// Start mock server
	mock := newMockServer(t)
	defer mock.Close()

	// Create client
	client, err := NewClient(mock.URL)
	if err != nil {
		t.Error(err)
	}

	// Test successful request
	collections, err := client.GetCollections(context.Background())
	require.NoError(t, err)
	require.NotNil(t, collections)

	// Verify the response matches our test data
	var expectedResp stac.CollectionsList
	err = json.Unmarshal(mock.Collections, &expectedResp)
	require.NoError(t, err)

	assert.Equal(t, len(expectedResp.Collections), len(collections.Collections))
	assert.Equal(t, expectedResp.Collections[0].Id, collections.Collections[0].Id)
	assert.Equal(t, expectedResp.Collections[1].Id, collections.Collections[1].Id)
}

func TestGetCollection(t *testing.T) {
	// Start mock server
	mock := newMockServer(t)
	defer mock.Close()

	// Create client
	client, err := NewClient(mock.URL)
	if err != nil {
		t.Error(err)
	}

	// Test successful request
	collection, err := client.GetCollection(context.Background(), "sentinel-2")
	require.NoError(t, err)
	require.NotNil(t, collection)

	// Verify the response matches our test data
	var expectedResp stac.Collection
	err = json.Unmarshal(mock.SingleCollection, &expectedResp)
	require.NoError(t, err)

	assert.Equal(t, expectedResp.Id, collection.Id)
	assert.Equal(t, expectedResp.Description, collection.Description)
	assert.Equal(t, expectedResp.Links[0].Href, collection.Links[0].Href)

	// Test non-existent collection
	collection, err = client.GetCollection(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.Nil(t, collection)
}

func TestSearchCollections(t *testing.T) {
	// Start mock server
	mock := newMockServer(t)
	defer mock.Close()

	// Create client
	client, err := NewClient(mock.URL)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name          string
		params        SearchCollectionsParams
		expectedError bool
	}{
		{
			name: "basic search",
			params: SearchCollectionsParams{
				Limit:  10,
				Query:  "sentinel",
				Fields: []string{"id", "description"},
			},
			expectedError: false,
		},
		{
			name: "search without query",
			params: SearchCollectionsParams{
				Limit:  5,
				Fields: []string{"id"},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collections, err := client.SearchCollections(context.Background(), tt.params)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, collections)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, collections)

				// Verify we got valid collections back
				assert.Greater(t, len(collections.Collections), 0)
				for _, collection := range collections.Collections {
					assert.NotEmpty(t, collection.Id)
					assert.NotEmpty(t, collection.Description)
				}
			}
		})
	}
}
