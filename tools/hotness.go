package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kkjdaniel/gogeek/v2"
	"github.com/kkjdaniel/gogeek/v2/hot"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func HotnessTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-hot",
		mcp.WithDescription("Find the current trending board game hotness on BoardGameGeek (BGG)"),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		hotItems, err := hot.Query(client, hot.ItemTypeBoardGame)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		if len(hotItems.Items) > 0 {
			out, err := json.Marshal(hotItems.Items)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
			}
			return mcp.NewToolResultText(string(out)), nil
		}

		return mcp.NewToolResultText("No hot games found"), nil
	}

	return tool, handler
}
