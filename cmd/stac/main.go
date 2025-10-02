package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

var (
	baseURLFlag = &cli.StringFlag{
		Name:     "url",
		Aliases:  []string{"u"},
		Usage:    "STAC API base URL",
		Required: true,
	}
	timeoutFlag = &cli.DurationFlag{
		Name:    "timeout",
		Aliases: []string{"t"},
		Usage:   "HTTP client timeout (e.g. 30s, 1m)",
		Value:   30 * time.Second,
	}
)

func main() {
	cmd := &cli.Command{
		Name:  "stac-cli",
		Usage: "Interact with STAC APIs",
		Flags: []cli.Flag{baseURLFlag, timeoutFlag},
		Commands: []*cli.Command{
			newCollectionsCommand(),
			newItemsCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func clientOptionsFromCommand(cmd *cli.Command) (string, time.Duration, error) {
	baseURL := cmd.String(baseURLFlag.Name)
	if baseURL == "" {
		return "", 0, fmt.Errorf("flag --url is required")
	}

	return baseURL, cmd.Duration(timeoutFlag.Name), nil
}
