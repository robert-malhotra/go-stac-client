package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	stac "github.com/planetlabs/go-stac"
	stacclient "github.com/robert-malhotra/go-stac-client/pkg/client"
	"github.com/urfave/cli/v3"
)

func newCollectionsCommand() *cli.Command {
	return &cli.Command{
		Name:  "collections",
		Usage: "Work with STAC collections",
		Commands: []*cli.Command{
			{
				Name:      "get",
				Usage:     "Fetch a collection by ID",
				ArgsUsage: "<collection-id>",
				Action:    getCollectionAction,
			},
			{
				Name:   "list",
				Usage:  "List all collections",
				Action: listCollectionsAction,
			},
		},
	}
}

func getCollectionAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 1 {
		return fmt.Errorf("expected 1 argument: collection id")
	}

	baseURL, timeout, err := clientOptionsFromCommand(cmd)
	if err != nil {
		return err
	}

	client, err := stacclient.NewClient(baseURL, stacclient.WithTimeout(timeout))
	if err != nil {
		return err
	}

	collection, err := client.GetCollection(ctx, cmd.Args().First())
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(newCollectionSummary(collection), "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(data))
	return nil
}

func listCollectionsAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 0 {
		return fmt.Errorf("no arguments expected")
	}

	baseURL, timeout, err := clientOptionsFromCommand(cmd)
	if err != nil {
		return err
	}

	client, err := stacclient.NewClient(baseURL, stacclient.WithTimeout(timeout))
	if err != nil {
		return err
	}

	seq := client.GetCollections(ctx)
	entries, err := collectForCLI(seq, func(c *stac.Collection) ([]byte, error) {
		return json.MarshalIndent(newCollectionSummary(c), "", "  ")
	})
	if err != nil {
		return err
	}

	return printJSONArray(entries)
}
