package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kkjdaniel/gogeek/v2"
	"github.com/kkjdaniel/gogeek/v2/thread"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ThreadDetailsTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-thread-details",
		mcp.WithDescription("Get full content of a specific BoardGameGeek forum thread, including all posts and replies. Use this after finding relevant threads with bgg-rules."),
		mcp.WithNumber("thread_id",
			mcp.Required(),
			mcp.Description("The BoardGameGeek thread ID to fetch"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		var threadID int
		var err error
		if idVal, ok := arguments["thread_id"]; ok && idVal != nil {
			switch v := idVal.(type) {
			case float64:
				threadID = int(v)
			case string:
				threadID, err = strconv.Atoi(v)
				if err != nil {
					return mcp.NewToolResultText("Invalid thread ID format"), nil
				}
			case int:
				threadID = v
			}
		} else {
			return mcp.NewToolResultText("thread_id parameter is required"), nil
		}
		threadDetail, err := thread.Query(client, threadID)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Failed to get thread details: %v", err)), nil
		}
		
		jsonResult, err := json.MarshalIndent(threadDetail, "", "  ")
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Failed to format result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return tool, handler
}