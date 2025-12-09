# go-stac-client

`go-stac-client` is a Go toolkit for working with [SpatioTemporal Asset Catalog (STAC)](https://stacspec.org/) APIs. It provides a composable HTTP client for the core STAC resources plus a lightweight CLI for quick exploration.

## Features

- Typed client for STAC `collections`, `items`, and `/search` endpoints with automatic pagination and request middleware hooks
- Streaming iteration built on Go 1.23 `iter.Seq2` so callers can stop early without buffering entire result sets
- CLI (`stac-cli`) built on urfave/cli v3 with support for fetching or listing collections/items and configurable timeouts
- Extensible pagination via pluggable `NextHandler` to accommodate custom link relations

## Installing the CLI

```bash
# Build and install the CLI binary locally
go install github.com/robert-malhotra/go-stac-client/cmd/stac@latest
```

Example usage:

```bash
stac-cli \
  --url https://example.com/stac \
  items list SENTINEL-2
```

Use `--timeout` (default `30s`) to adjust the HTTP timeout for all commands.

## Library Quick Start

```go
package main

import (
    "context"
    "fmt"
    "iter"

    stac "github.com/planetlabs/go-stac"
    stacclient "github.com/robert-malhotra/go-stac-client/pkg/client"
)

func main() {
    cli, err := stacclient.NewClient("https://earth-search.aws.element84.com/v1")
    if err != nil {
        panic(err)
    }

    // Stream collections
    for col, err := range cli.GetCollections(context.Background()) {
        if err != nil {
            panic(err)
        }
        fmt.Println(col.Id)
    }

    // Collect search results (GET)
    params := stacclient.SearchParams{
        Collections: []string{"sentinel-2-l2a"},
        Bbox:        []float64{-123.3, 45.2, -122.5, 46.0},
        Limit:       10,
    }
    items, err := collect(cli.SearchSimple(context.Background(), params))
    if err != nil {
        panic(err)
    }
    fmt.Printf("retrieved %d items\n", len(items))
}

// helper to exhaust an iterator
func collect(seq iter.Seq2[*stac.Item, error]) ([]*stac.Item, error) {
    var (
        out []*stac.Item
        outErr error
    )
    seq(func(it *stac.Item, err error) bool {
        if err != nil {
            outErr = err
            return false
        }
        out = append(out, it)
        return true
    })
    return out, outErr
}
```

## Testing

```bash
# Unit tests
go test ./...

# Skip live network calls
go test ./... -short
```

Two integration tests hit public STAC endpoints (Copernicus Dataspace and NASA ASF). They are skipped automatically when `-short` is provided.

## License

[MIT](LICENSE)
