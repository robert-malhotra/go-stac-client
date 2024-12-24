// File: cmd/collections/collections.go
package collections

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go-stac-client/pkg/stac"

	"github.com/spf13/cobra"
)

var (
	limit  int
	query  string
	fields []string
	output string
)

// Command returns the collections command
func NewCollectionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collections",
		Short: "Interact with STAC collections",
		Long:  `List, search, and retrieve information about STAC collections.`,
	}

	// Add subcommands
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newSearchCmd())

	return cmd
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL, _ := cmd.Flags().GetString("url")
			client := stac.NewClient(baseURL)
			collections, err := client.GetCollections(context.Background())
			if err != nil {
				return fmt.Errorf("error getting collections: %w", err)
			}

			return outputResults(collections)
		},
	}
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [collection-id]",
		Short: "Get a specific collection by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL, _ := cmd.Flags().GetString("url")
			client := stac.NewClient(baseURL)
			collection, err := client.GetCollection(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("error getting collection: %w", err)
			}

			return outputResults(collection)
		},
	}
}

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search collections with parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL, _ := cmd.Flags().GetString("url")
			client := stac.NewClient(baseURL)
			params := stac.SearchCollectionsParams{
				Limit:  limit,
				Query:  query,
				Fields: fields,
			}

			collections, err := client.SearchCollections(context.Background(), params)
			if err != nil {
				return fmt.Errorf("error searching collections: %w", err)
			}

			return outputResults(collections)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results to return")
	cmd.Flags().StringVar(&query, "query", "", "Search query string")
	cmd.Flags().StringSliceVar(&fields, "fields", []string{}, "Comma-separated list of fields to return")

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
	case *stac.CollectionsResponse:
		fmt.Printf("%-36s %-40s %-s\n", "ID", "TITLE", "DESCRIPTION")
		fmt.Println(strings.Repeat("-", 100))
		for _, collection := range v.Collections {
			fmt.Printf("%-36s %-40s %-s\n",
				truncateString(collection.ID, 36),
				truncateString(collection.Title, 40),
				truncateString(collection.Description, 50))
		}
	case *stac.Collection:
		fmt.Printf("ID: %s\n", v.ID)
		fmt.Printf("Title: %s\n", v.Title)
		fmt.Printf("Description: %s\n", v.Description)
		fmt.Printf("License: %s\n", v.License)
		if len(v.Keywords) > 0 {
			fmt.Printf("Keywords: %s\n", strings.Join(v.Keywords, ", "))
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
