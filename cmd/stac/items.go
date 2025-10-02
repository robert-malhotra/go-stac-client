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

func newItemsCommand() *cli.Command {
	return &cli.Command{
		Name:  "items",
		Usage: "Work with STAC items",
		Commands: []*cli.Command{
			{
				Name:      "get",
				Usage:     "Fetch an item by collection and ID",
				ArgsUsage: "<collection-id> <item-id>",
				Action:    getItemAction,
			},
			{
				Name:      "list",
				Usage:     "List items in a collection",
				ArgsUsage: "<collection-id>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "interactive",
						Aliases: []string{"i"},
						Usage:   "Prompt between batches of results",
					},
				},
				Action: listItemsAction,
			},
		},
	}
}

func getItemAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 2 {
		return fmt.Errorf("expected 2 arguments: collection id and item id")
	}

	baseURL, timeout, err := clientOptionsFromCommand(cmd)
	if err != nil {
		return err
	}

	client, err := stacclient.NewClient(baseURL, stacclient.WithTimeout(timeout))
	if err != nil {
		return err
	}

	item, err := client.GetItem(ctx, cmd.Args().Get(0), cmd.Args().Get(1))
	if err != nil {
		return err
	}

	summary, err := newItemSummary(item)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(data))
	return nil
}

func listItemsAction(ctx context.Context, cmd *cli.Command) error {
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

	seq := client.GetItems(ctx, cmd.Args().First())
	marshal := func(item *stac.Item) ([]byte, error) {
		summary, err := newItemSummary(item)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(summary, "", "  ")
	}

	if cmd.Bool("interactive") {
		return printJSONArrayInteractive(seq, marshal)
	}

	entries, err := collectForCLI(seq, marshal)
	if err != nil {
		return err
	}

	return printJSONArray(entries)
}
