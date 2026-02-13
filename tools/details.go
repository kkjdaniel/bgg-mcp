package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kkjdaniel/gogeek/v2"
	"github.com/kkjdaniel/gogeek/v2/thing"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func DetailsTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-details",
		mcp.WithDescription("Get detailed information about board games on BoardGameGeek (BGG) including description, mechanics, categories, player count, playtime, complexity, and ratings. Use this tool to deep dive into games found via other tools (e.g. after getting collection results or search results that only return basic info). Use 'name' for a single game lookup by name, 'id' for a single game lookup by BGG ID, or 'ids' to fetch multiple games at once (up to 20). Only provide one of these parameters."),
		mcp.WithString("name",
			mcp.Description("The name of the board game to look up. Use this when you only have the game's name."),
		),
		mcp.WithNumber("id",
			mcp.Description("The BoardGameGeek ID of a single board game. Preferred over 'name' when the ID is already known."),
		),
		mcp.WithArray("ids",
			mcp.Description("Array of BoardGameGeek IDs for fetching multiple games in a single request (maximum 20). Use this instead of 'id' when you need details for more than one game."),
			mcp.WithNumberItems(),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()

		var gameIDs []int
		var err error

		if idsVal, ok := arguments["ids"]; ok && idsVal != nil {
			idsArray, ok := idsVal.([]interface{})
			if !ok {
				return mcp.NewToolResultText("Invalid IDs format - must be an array"), nil
			}
			
			if len(idsArray) > 20 {
				return mcp.NewToolResultText("Too many IDs provided. Maximum 20 IDs per request."), nil
			}
			
			for _, idVal := range idsArray {
				var gameID int
				switch v := idVal.(type) {
				case float64:
					gameID = int(v)
				case string:
					gameID, err = strconv.Atoi(v)
					if err != nil {
						return mcp.NewToolResultText(fmt.Sprintf("Invalid ID format: %s", v)), nil
					}
				default:
					return mcp.NewToolResultText("Invalid ID type in array"), nil
				}
				gameIDs = append(gameIDs, gameID)
			}
		} else if idVal, ok := arguments["id"]; ok && idVal != nil {
			var gameID int
			switch v := idVal.(type) {
			case float64:
				gameID = int(v)
			case string:
				gameID, err = strconv.Atoi(v)
				if err != nil {
					return mcp.NewToolResultText("Invalid ID format"), nil
				}
			default:
				return mcp.NewToolResultText("Invalid ID type"), nil
			}
			gameIDs = []int{gameID}
		} else if nameVal, ok := arguments["name"]; ok && nameVal != nil {
			name := nameVal.(string)
			bestMatch, err := findBestGameMatch(client, name)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Failed to find game: %v", err)), nil
			}
			gameIDs = []int{bestMatch.ID}
		} else {
			return mcp.NewToolResultText("Either 'name', 'id', or 'ids' parameter must be provided"), nil
		}

		things, err := thing.Query(client, gameIDs)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		if len(things.Items) > 0 {
			var out []byte
			var err error

			if len(gameIDs) == 1 {
				essentialInfo := extractEssentialInfo(things.Items[0])
				out, err = json.Marshal(essentialInfo)
			} else {
				essentialInfo := extractEssentialInfoList(things.Items)
				out, err = json.Marshal(essentialInfo)
			}

			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
			}
			return mcp.NewToolResultText(string(out)), nil
		}

		return mcp.NewToolResultText("No query results found"), nil
	}

	return tool, handler
}
