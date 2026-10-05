package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/user"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func UserTool(client *gogeek.Client, selfUsername string) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-user",
		mcp.WithDescription("Find details about a specific user on BoardGameGeek (BGG)"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the BoardGameGeek (BGG) user. When the user refers to themselves (me, my, I), use 'SELF' as the value."),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		name := arguments["username"].(string)

		if name == "SELF" {
			if selfUsername == "" {
				return mcp.NewToolResultText(selfUsernameMissingMessage), nil
			}
			name = selfUsername
		}

		userDetails, err := user.Query(ctx, client, name)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		out, err := json.Marshal(userDetails)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil

	}

	return tool, handler
}
