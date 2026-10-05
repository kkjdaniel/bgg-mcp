package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/collection"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ownedCollectionContents(ctx context.Context, client *gogeek.Client, uri string, username string) ([]mcp.ResourceContents, error) {
	result, err := collection.Query(ctx, client, username, collection.WithOwned(true))
	if err != nil {
		return nil, fmt.Errorf("error fetching collection: %v", err)
	}

	if len(result.Items) == 0 {
		return jsonContents(uri, "[]"), nil
	}

	out, err := json.Marshal(result.Items)
	if err != nil {
		return nil, fmt.Errorf("error formatting results: %v", err)
	}

	return jsonContents(uri, string(out)), nil
}

// MyCollectionResource exposes the collection of the configured user. It should
// only be registered when a username is configured.
func MyCollectionResource(client *gogeek.Client, username string) (mcp.Resource, server.ResourceHandlerFunc) {
	resource := mcp.NewResource(
		"bgg://my-collection",
		"My BGG Collection",
		mcp.WithResourceDescription(fmt.Sprintf("Your BoardGameGeek collection (user: %s). Shows all owned games with their ratings, play counts, and status.", username)),
		mcp.WithMIMEType("application/json"),
	)

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		return ownedCollectionContents(ctx, client, "bgg://my-collection", username)
	}

	return resource, handler
}
