# go-stac-client

An idiomatic Go client for interacting with STAC APIs. It reuses the [planetlabs/go-stac](https://github.com/planetlabs/go-stac) data models and provides a fluent query builder backed by the OGC filter primitives.

## Features

- Configurable HTTP client with retry hooks and request options
- Collection, item, and search services backed by `planetlabs/go-stac` models
- Lazy pagination using Go 1.22+ iterator sequences (`iter.Seq2`)
- Query builder that emits CQL2 JSON via the `github.com/planetlabs/go-ogc/filter` package
- Simple authentication transports for API key and bearer token flows

## Quick Start

```go
package main

import (
    "context"
    "log"

    stacclient "github.com/example/go-stac-client/client"
    "github.com/example/go-stac-client/query"
)

func main() {
    client, err := stacclient.New(
        stacclient.WithBaseURL("https://planetarycomputer.microsoft.com/api/stac/v1"),
    )
    if err != nil {
        log.Fatal(err)
    }

    params := stacclient.SearchParams{Limit: 10}
    params.Filter = query.NewBuilder().
        Where(query.Property("eo:cloud_cover").Lte(10)).
        Filter()

    for item, err := range client.Search().Query(context.Background(), params) {
        if err != nil {
            log.Fatal(err)
        }
        log.Println("found", item.Id)
    }
}
```

## Manual Pagination

If you need to inspect pagination metadata or drive custom paging, use the `GetPage` helpers:

```go
page, err := client.Search().GetPage(ctx, params)
if err != nil {
    return err
}
for _, item := range page.Items {
    // process each *stac.Item
}
nextToken := page.NextToken()
```

## Development

```bash
go test ./...
```
