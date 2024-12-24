// File: cmd/init.go
package cmd

import (
	"go-stac-client/cmd/collections"
	"go-stac-client/cmd/items"
)

func init() {
	// Add collections command
	rootCmd.AddCommand(collections.NewCollectionsCmd())
	rootCmd.AddCommand(items.NewItemsCmd())
}
