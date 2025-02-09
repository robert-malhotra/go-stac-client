package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	stac_client "github.com/robert-malhotra/go-stac-client/pkg/client"

	"github.com/spf13/cobra"
)

var (
	limit  int
	query  string
	fields []string
	format string
)

var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Work with STAC collections",
	Long:  `List and retrieve STAC collections from the API.`,
}

var listCollectionsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all collections",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := stac_client.NewClient(baseURL)
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		params := stac_client.SearchCollectionsParams{
			Limit:  limit,
			Query:  query,
			Fields: fields,
		}

		collections, err := client.SearchCollections(context.Background(), params)
		if err != nil {
			return fmt.Errorf("error getting collections: %w", err)
		}

		switch format {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(collections)
		case "text":
			for _, c := range collections.Collections {
				fmt.Printf("ID: %s\n", c.Id)
				fmt.Printf("Title: %s\n", c.Title)
				fmt.Printf("Description: %s\n", c.Description)
				fmt.Println("---")
			}
		default:
			return fmt.Errorf("unknown format: %s", format)
		}

		return nil
	},
}

var getCollectionCmd = &cobra.Command{
	Use:   "get [collection-id]",
	Short: "Get a specific collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := stac_client.NewClient(baseURL)
		if err != nil {
			return fmt.Errorf("error creating client: %w", err)
		}

		collection, err := client.GetCollection(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("error getting collection: %w", err)
		}

		switch format {
		case "json":
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(collection)
		case "text":
			fmt.Printf("ID: %s\n", collection.Id)
			fmt.Printf("Title: %s\n", collection.Title)
			fmt.Printf("Description: %s\n", collection.Description)
			fmt.Printf("License: %s\n", collection.License)
			if len(collection.Keywords) > 0 {
				fmt.Printf("Keywords: %v\n", collection.Keywords)
			}
		default:
			return fmt.Errorf("unknown format: %s", format)
		}

		return nil
	},
}

func init() {
	// Add collections command to root
	rootCmd.AddCommand(collectionsCmd)

	// Add subcommands to collections
	collectionsCmd.AddCommand(listCollectionsCmd)
	collectionsCmd.AddCommand(getCollectionCmd)

	// Add flags to list command
	listCollectionsCmd.Flags().IntVar(&limit, "limit", 0, "Maximum number of collections to return")
	listCollectionsCmd.Flags().StringVar(&query, "query", "", "Query string to filter collections")
	listCollectionsCmd.Flags().StringSliceVar(&fields, "fields", nil, "Fields to include in the response")
	listCollectionsCmd.Flags().StringVar(&format, "format", "json", "Output format (json or text)")

	// Add flags to get command
	getCollectionCmd.Flags().StringVar(&format, "format", "json", "Output format (json or text)")
}
