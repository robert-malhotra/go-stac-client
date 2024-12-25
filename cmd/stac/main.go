package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stac",
	Short: "A STAC API client CLI",
	Long: `A command line interface for interacting with STAC APIs.
Supports collection and item operations with various filtering options.`,
}

var baseURL string

func init() {
	rootCmd.PersistentFlags().StringVar(&baseURL, "url", "", "Base URL of the STAC API (required)")
	rootCmd.MarkPersistentFlagRequired("url")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
