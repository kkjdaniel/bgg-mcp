package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kkjdaniel/bgg-mcp/tools"
	"github.com/kkjdaniel/gogeek/v3"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func HotnessResource(client *gogeek.Client) (mcp.Resource, server.ResourceHandlerFunc) {
	resource := mcp.NewResource(
		"bgg://hotness",
		"Current BGG Hotness List",
		mcp.WithResourceDescription("The current list of hot board games on BoardGameGeek, updated regularly by the BGG community"),
		mcp.WithMIMEType("application/json"),
	)

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		hotGames, err := tools.FetchHotGames(ctx, client)
		if err != nil {
			return nil, fmt.Errorf("error fetching hotness list: %v", err)
		}

		out, err := json.Marshal(hotGames)
		if err != nil {
			return nil, fmt.Errorf("error formatting results: %v", err)
		}

		return jsonContents("bgg://hotness", string(out)), nil
	}

	return resource, handler
}
