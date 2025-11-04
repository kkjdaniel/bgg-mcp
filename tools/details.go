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
		mcp.WithDescription("Find the details about a specific board game on BoardGameGeek (BGG)"),
		mcp.WithString("name",
			mcp.Description("The name of the board game"),
		),
		mcp.WithNumber("id",
			mcp.Description("The BoardGameGeek ID of the board game"),
		),
		mcp.WithArray("ids",
			mcp.Description("Array of BoardGameGeek IDs to get details for multiple games at once (maximum 20 IDs per request)"),
		),
		mcp.WithBoolean("full_details",
			mcp.Description("Return the complete BGG API response instead of essential info. WARNING: This returns significantly more data and can overload AI context windows. ONLY set this to true if the user explicitly requests 'full details', 'complete data', or similar. Default behavior returns essential info which is sufficient for most use cases."),
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
			fullDetails := false
			if fd, ok := arguments["full_details"].(bool); ok {
				fullDetails = fd
			}

			var out []byte
			var err error
			
			if len(gameIDs) == 1 {
				if fullDetails {
					out, err = json.Marshal(things.Items[0])
				} else {
					essentialInfo := extractEssentialInfo(things.Items[0])
					out, err = json.Marshal(essentialInfo)
				}
			} else {
				if fullDetails {
					out, err = json.Marshal(things.Items)
				} else {
					essentialInfo := extractEssentialInfoList(things.Items)
					out, err = json.Marshal(essentialInfo)
				}
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
