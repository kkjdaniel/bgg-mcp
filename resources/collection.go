package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kkjdaniel/gogeek/v2"
	"github.com/kkjdaniel/gogeek/v2/collection"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func MyCollectionResource(client *gogeek.Client) (mcp.Resource, server.ResourceHandlerFunc) {
	username := os.Getenv("BGG_USERNAME")

	var description string
	if username != "" {
		description = fmt.Sprintf("Your BoardGameGeek collection (user: %s). Shows all owned games with their ratings, play counts, and status.", username)
	} else {
		description = "Your BoardGameGeek collection. Requires BGG_USERNAME environment variable to be set."
	}

	resource := mcp.NewResource(
		"bgg://my-collection",
		"My BGG Collection",
		mcp.WithResourceDescription(description),
		mcp.WithMIMEType("application/json"),
	)

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		if username == "" {
			return nil, fmt.Errorf("BGG_USERNAME environment variable not set")
		}

		result, err := collection.Query(client, username, collection.WithOwned(true))
		if err != nil {
			return nil, fmt.Errorf("error fetching collection: %v", err)
		}

		if len(result.Items) == 0 {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      "bgg://my-collection",
					MIMEType: "application/json",
					Text:     "[]",
				},
			}, nil
		}

		out, err := json.Marshal(result.Items)
		if err != nil {
			return nil, fmt.Errorf("error formatting results: %v", err)
		}

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      "bgg://my-collection",
				MIMEType: "application/json",
				Text:     string(out),
			},
		}, nil
	}

	return resource, handler
}
