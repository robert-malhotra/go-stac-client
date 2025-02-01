package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/planetlabs/go-stac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockServer(t *testing.T) *httptest.Server {
	itemsData := loadTestData(t, "items.json")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/collections/SENTINEL-1/items":
			// For filters, parse and filter the data
			if filter := r.URL.Query().Get("filter"); filter != "" {
				// Parse the original data
				var fullResponse map[string]interface{}
				err := json.Unmarshal(itemsData, &fullResponse)
				require.NoError(t, err)

				// Filter the features
				features := fullResponse["features"].([]interface{})
				filteredFeatures := make([]interface{}, 0)

				for _, feature := range features {
					f := feature.(map[string]interface{})
					props := f["properties"].(map[string]interface{})

					// Apply the filter
					if filter == "productType = 'AUX_CAL'" && props["productType"] == "AUX_CAL" {
						filteredFeatures = append(filteredFeatures, feature)
					} else if filter == "productType = 'AUX_GNSSRD'" && props["productType"] == "AUX_GNSSRD" {
						filteredFeatures = append(filteredFeatures, feature)
					}
				}

				fullResponse["features"] = filteredFeatures
				json.NewEncoder(w).Encode(fullResponse)
				return
			}

			// Default: return all items
			w.Write(itemsData)

		case "/collections/SENTINEL-1/items/S1A_AUX_CAL_V20140908T000000_G20240327T101157.SAFE":
			// Parse the full response
			var fullResponse map[string]interface{}
			err := json.Unmarshal(itemsData, &fullResponse)
			require.NoError(t, err)

			// Find the requested item
			features := fullResponse["features"].([]interface{})
			for _, feature := range features {
				f := feature.(map[string]interface{})
				if f["id"] == "S1A_AUX_CAL_V20140908T000000_G20240327T101157.SAFE" {
					json.NewEncoder(w).Encode(f)
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return server
}

func TestGetItems(t *testing.T) {
	server := createMockServer(t)
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	tests := []struct {
		name          string
		collectionID  string
		params        *ItemsParams
		expectError   bool
		validateItems func(*testing.T, *stac.ItemsList)
	}{
		{
			name:         "basic request",
			collectionID: "SENTINEL-1",
			params:       nil,
			expectError:  false,
			validateItems: func(t *testing.T, response *stac.ItemsList) {
				assert.NotEmpty(t, response.Items)
				assert.Equal(t, "FeatureCollection", response.Type)

				// Validate first item
				item := response.Items[0]
				assert.Equal(t, "S1A_AUX_CAL_V20140908T000000_G20240327T101157.SAFE", item.Id)
				assert.Equal(t, "SENTINEL-1", item.Collection)
				assert.Contains(t, item.Properties, "productType")
				assert.Contains(t, item.Assets, "PRODUCT")

				// Validate STAC extensions
				// assert.Contains(t, item.Extensions, "https://stac-extensions.github.io/alternate-assets/v1.1.0/schema.json")
				// assert.Contains(t, item.Extensions, "https://stac-extensions.github.io/storage/v1.0.0/schema.json")
			},
		},
		{
			name:         "filter AUX_GNSSRD products",
			collectionID: "SENTINEL-1",
			params: &ItemsParams{
				Filter: "productType = 'AUX_GNSSRD'",
			},
			expectError: false,
			validateItems: func(t *testing.T, response *stac.ItemsList) {
				assert.NotEmpty(t, response.Items)
				for _, item := range response.Items {
					assert.Equal(t, "AUX_GNSSRD", item.Properties["productType"])
				}
			},
		},
		{
			name:         "invalid collection",
			collectionID: "INVALID",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.GetItems(context.Background(), tt.collectionID, tt.params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, response)
			if tt.validateItems != nil {
				tt.validateItems(t, response)
			}
		})
	}
}

func TestGetItem(t *testing.T) {
	server := createMockServer(t)
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	// Test getting a specific item
	itemID := "S1A_AUX_CAL_V20140908T000000_G20240327T101157.SAFE"
	item, err := client.GetItem(context.Background(), "SENTINEL-1", itemID)
	require.NoError(t, err)
	require.NotNil(t, item)

	// Validate item fields
	assert.Equal(t, itemID, item.Id)
	assert.Equal(t, "SENTINEL-1", item.Collection)
	assert.NotNil(t, item.Properties)
	assert.Contains(t, item.Properties, "productType")
	assert.Contains(t, item.Assets, "PRODUCT")

	// Validate STAC extensions
	// assert.Contains(t, item.Extensions, "https://stac-extensions.github.io/alternate-assets/v1.1.0/schema.json")
	// assert.Contains(t, item.Extensions, "https://stac-extensions.github.io/storage/v1.0.0/schema.json")

	// Test non-existent item
	item, err = client.GetItem(context.Background(), "SENTINEL-1", "non-existent")
	assert.Error(t, err)
	assert.Nil(t, item)
}

func TestLiveESA(t *testing.T) {

	baseURL := "https://catalogue.dataspace.copernicus.eu/stac"
	client, err := NewClient(baseURL)
	require.NoError(t, err)

	r, err := client.GetItems(context.Background(), "SENTINEL-1", nil)
	require.NoError(t, err)

	if len(r.Items) == 0 {
		t.Error("received 0 items back")
	}
	t.Log(r)

	_, err = client.GetNextItems(context.Background(), r)
	require.NoError(t, err)

}
