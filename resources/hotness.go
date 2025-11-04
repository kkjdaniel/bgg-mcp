package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kkjdaniel/gogeek/v2"
	"github.com/kkjdaniel/gogeek/v2/hot"
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
		hotItems, err := hot.Query(client, hot.ItemTypeBoardGame)
		if err != nil {
			return nil, fmt.Errorf("error fetching hotness list: %v", err)
		}

		if len(hotItems.Items) == 0 {
			return []mcp.ResourceContents{
				&mcp.TextResourceContents{
					URI:      "bgg://hotness",
					MIMEType: "application/json",
					Text:     "[]",
				},
			}, nil
		}

		out, err := json.Marshal(hotItems.Items)
		if err != nil {
			return nil, fmt.Errorf("error formatting results: %v", err)
		}

		return []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      "bgg://hotness",
				MIMEType: "application/json",
				Text:     string(out),
			},
		}, nil
	}

	return resource, handler
}
