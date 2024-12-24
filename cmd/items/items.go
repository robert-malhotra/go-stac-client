// File: cmd/items/items.go
package items

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go-stac-client/pkg/stac"

	"github.com/spf13/cobra"
)

var (
	limit       int
	bbox        []float64
	datetime    string
	collections []string
	output      string
	filter      string
)

func NewItemsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "items",
		Short: "Interact with STAC items",
		Long:  `List, search, and retrieve information about STAC items.`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newSearchCmd())

	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [collection-id]",
		Short: "List items in a collection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL, _ := cmd.Flags().GetString("url")
			client := stac.NewClient(baseURL)
			items, err := client.GetCollectionItems(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("error getting items: %w", err)
			}

			return outputResults(items)
		},
	}

	return cmd
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [collection-id] [item-id]",
		Short: "Get a specific item from a collection",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL, _ := cmd.Flags().GetString("url")
			client := stac.NewClient(baseURL)
			item, err := client.GetItem(context.Background(), args[0], args[1])
			if err != nil {
				return fmt.Errorf("error getting item: %w", err)
			}

			return outputResults(item)
		},
	}
}

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search items across collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL, _ := cmd.Flags().GetString("url")
			client := stac.NewClient(baseURL)

			params := stac.SearchItemsParams{
				Collections: collections,
				BBox:        bbox,
				Datetime:    datetime,
				Limit:       limit,
			}

			// Parse CQL2-JSON filter if provided
			if filter != "" {
				var filterJson map[string]interface{}
				if err := json.Unmarshal([]byte(filter), &filterJson); err != nil {
					return fmt.Errorf("error parsing filter JSON: %w", err)
				}
				params.Filter = filterJson
			}

			items, err := client.SearchItems(context.Background(), params)
			if err != nil {
				return fmt.Errorf("error searching items: %w", err)
			}

			return outputResults(items)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results to return")
	cmd.Flags().Float64SliceVar(&bbox, "bbox", nil, "Bounding box coordinates (minX,minY,maxX,maxY)")
	cmd.Flags().StringVar(&datetime, "datetime", "", "Datetime filter (e.g., 2024-01-01/2024-12-31)")
	cmd.Flags().StringSliceVar(&collections, "collections", nil, "Collection IDs to search within")
	cmd.Flags().StringVar(&filter, "filter", "", "CQL2-JSON filter")

	return cmd
}

func outputResults(data interface{}) error {
	var err error
	switch output {
	case "json":
		err = outputJSON(data)
	case "table":
		err = outputTable(data)
	default:
		err = outputJSON(data)
	}
	return err
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputTable(data interface{}) error {
	switch v := data.(type) {
	case *stac.ItemsResponse:
		fmt.Printf("%-36s %-24s %-40s\n", "ID", "DATETIME", "TITLE")
		fmt.Println(strings.Repeat("-", 100))
		for _, item := range v.Features {
			datetime := ""
			if dt, ok := item.Properties["datetime"].(string); ok {
				if t, err := time.Parse(time.RFC3339, dt); err == nil {
					datetime = t.Format("2006-01-02 15:04:05")
				}
			}
			title := ""
			if t, ok := item.Properties["title"].(string); ok {
				title = t
			}
			fmt.Printf("%-36s %-24s %-40s\n",
				truncateString(item.ID, 36),
				truncateString(datetime, 24),
				truncateString(title, 40))
		}
	case *stac.Item:
		fmt.Printf("ID: %s\n", v.ID)
		if dt, ok := v.Properties["datetime"].(string); ok {
			fmt.Printf("Datetime: %s\n", dt)
		}
		if title, ok := v.Properties["title"].(string); ok {
			fmt.Printf("Title: %s\n", title)
		}
		fmt.Println("\nAssets:")
		for name, asset := range v.Assets {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    Href: %s\n", asset.Href)
			if asset.Type != "" {
				fmt.Printf("    Type: %s\n", asset.Type)
			}
			if len(asset.Roles) > 0 {
				fmt.Printf("    Roles: %s\n", strings.Join(asset.Roles, ", "))
			}
		}
	default:
		return fmt.Errorf("unsupported data type for table output")
	}
	return nil
}

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
