package main

import (
	"context"
	"encoding/json"
	"fmt"
	stac_client "go-stac-client/pkg/client"
	"os"
	"strings"
	"time"

	stac "github.com/planetlabs/go-stac"
	"github.com/spf13/cobra"
)

var (
	itemLimit  int
	itemFilter string
	datetime   string
	dateEnd    string
	bbox       []float64
	itemFields []string
	sortBy     []string
	dumpAll    bool
)

var itemsCmd = &cobra.Command{
	Use:   "items",
	Short: "Work with STAC items",
	Long:  `List and retrieve STAC items from collections.`,
}

var listItemsCmd = &cobra.Command{
	Use:   "list [collection-id]",
	Short: "List items in a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := stac_client.NewClient(baseURL)
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		params := &stac_client.ItemsParams{
			Filter: itemFilter,
			Fields: itemFields,
			SortBy: sortBy,
		}

		if itemLimit > 0 {
			limit := itemLimit
			params.Limit = &limit
		}

		if datetime != "" {
			t, err := time.Parse(time.RFC3339, datetime)
			if err != nil {
				return fmt.Errorf("invalid datetime format: %w", err)
			}
			params.DateTime = &t

			if dateEnd != "" {
				end, err := time.Parse(time.RFC3339, dateEnd)
				if err != nil {
					return fmt.Errorf("invalid end datetime format: %w", err)
				}
				params.DateEnd = &end
			}
		}

		if len(bbox) > 0 {
			params.BBox = bbox
		}

		ctx := context.Background()

		// Get first page
		items, err := client.GetItems(ctx, args[0], params)
		if err != nil {
			return fmt.Errorf("error getting items: %w", err)
		}

		if !dumpAll {
			// Output single page
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(items)
		}

		// Collect all items into one FeatureCollection
		allItems := &stac.ItemsList{
			Type:  items.Type,
			Items: make([]*stac.Item, 0),
		}
		allItems.Items = append(allItems.Items, items.Items...)

		for {

			if !hasNext(items.Links) {
				break
			}

			items, err = client.GetNextItems(ctx, items)
			if err != nil {
				return fmt.Errorf("error getting next page: %w", err)
			}
			allItems.Items = append(allItems.Items, items.Items...)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allItems)
	},
}

var getItemCmd = &cobra.Command{
	Use:   "get [collection-id] [item-id]",
	Short: "Get a specific item",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := stac_client.NewClient(baseURL)
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		item, err := client.GetItem(context.Background(), args[0], args[1])
		if err != nil {
			return fmt.Errorf("error getting item: %w", err)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(item)
	},
}

func hasNext(links []*stac.Link) bool {
	for _, link := range links {
		if strings.ToLower(link.Rel) == "next" {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(itemsCmd)
	itemsCmd.AddCommand(listItemsCmd)
	itemsCmd.AddCommand(getItemCmd)

	listItemsCmd.Flags().IntVar(&itemLimit, "limit", 100, "Maximum number of items to return per page")
	listItemsCmd.Flags().StringVar(&itemFilter, "filter", "", "CQL filter to apply")
	listItemsCmd.Flags().StringVar(&datetime, "datetime", "", "Datetime filter (RFC3339 format)")
	listItemsCmd.Flags().StringVar(&dateEnd, "datetime-end", "", "End datetime for range (RFC3339 format)")
	listItemsCmd.Flags().Float64SliceVar(&bbox, "bbox", nil, "Bounding box (minx,miny,maxx,maxy)")
	listItemsCmd.Flags().StringSliceVar(&itemFields, "fields", nil, "Fields to include in the response")
	listItemsCmd.Flags().StringSliceVar(&sortBy, "sort-by", nil, "Fields to sort by")
	listItemsCmd.Flags().BoolVar(&dumpAll, "all", false, "Collect all pages into a single FeatureCollection")
}
