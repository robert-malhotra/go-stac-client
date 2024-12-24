// File: cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	baseURL string
	output  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stac",
	Short: "A CLI for interacting with STAC APIs",
	Long: `A command line interface for interacting with Spatiotemporal Asset Catalog (STAC) APIs.
This CLI allows you to list, search, and retrieve collection information.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&baseURL, "url", "", "STAC API base URL (required)")
	rootCmd.PersistentFlags().StringVar(&output, "output", "json", "Output format (json or table)")
	rootCmd.MarkPersistentFlagRequired("url")
}
