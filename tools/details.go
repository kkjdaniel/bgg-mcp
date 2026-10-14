package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/thing"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// commentsPageSize is how many comments the comments option fetches per game:
// enough to get a sense of opinion without swamping the response.
const commentsPageSize = 50

func DetailsTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-details",
		mcp.WithDescription("Get detailed information about board games on BoardGameGeek (BGG) including description, mechanics, categories, player count, the player counts the BGG community votes best and recommended ('best_players', 'recommended_players'), playtime, complexity, ratings, BGG rank, language dependence, and links between expansions and base games ('expands'; a single-game lookup also lists the game's 'expansions'). Use this tool to deep dive into games found via other tools (e.g. after getting collection results or search results that only return basic info). Use 'name' for a single game lookup by name, 'id' for a single game lookup by BGG ID, or 'ids' to fetch multiple games at once (up to 20). Only provide one of these parameters."),
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
		mcp.WithBoolean("videos",
			mcp.Description("Include the community-submitted videos BGG links to each game, such as reviews, playthroughs and rules explanations. Off by default, because the list lengthens a response that otherwise carries only essential information. The 'total_on_bgg' figure returned alongside counts what BGG holds for the game and is routinely far larger than the list itself."),
		),
		mcp.WithBoolean("comments",
			mcp.Description(fmt.Sprintf("Include a page of the comments BGG users have left on each game, with the rating each gave it. Use this for what players think of a game. Off by default; up to %d comments per game, and 'total_on_bgg' counts every comment BGG holds.", commentsPageSize)),
		),
		mcp.WithBoolean("marketplace",
			mcp.Description("Include the copies of each game for sale on the BGG marketplace, with price, currency, condition and a link to each listing. Use this for what a game sells for second-hand. Off by default."),
		),
		mcp.WithBoolean("versions",
			mcp.Description("Include each game's published editions (printings and localisations) with their year, publishers and languages. Use this for questions about editions or which language versions exist. Off by default, because a long-lived game can have hundreds."),
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
			bestMatch, err := findBestGameMatch(ctx, client, name)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Failed to find game: %v", err)), nil
			}
			gameIDs = []int{bestMatch.ID}
		} else {
			return mcp.NewToolResultText("Either 'name', 'id', or 'ids' parameter must be provided"), nil
		}

		var options []thing.Option
		if withVideos, ok := arguments["videos"].(bool); ok && withVideos {
			options = append(options, thing.WithVideos())
		}
		withComments, _ := arguments["comments"].(bool)
		if withComments {
			options = append(options, thing.WithComments(), thing.WithPageSize(commentsPageSize))
		}
		withMarketplace, _ := arguments["marketplace"].(bool)
		if withMarketplace {
			options = append(options, thing.WithMarketplace())
		}
		withVersions, _ := arguments["versions"].(bool)
		if withVersions {
			options = append(options, thing.WithVersions())
		}
		// A game nobody sells, or with no editions listed, looks the same as one
		// queried without the option, so say so rather than leave the field out.
		markEmpty := func(info *EssentialGameInfo) {
			if withMarketplace && info.Marketplace == nil {
				info.Marketplace = &GameMarketplace{Listings: []MarketplaceListing{}}
			}
			if withVersions && info.Versions == nil {
				info.Versions = &GameVersions{Versions: []GameVersion{}}
			}
		}

		things, err := thing.Query(ctx, client, gameIDs, options...)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		if len(things.Items) > 0 {
			var out []byte
			var err error

			if len(gameIDs) == 1 {
				essentialInfo := extractEssentialInfo(things.Items[0])
				markEmpty(&essentialInfo)
				out, err = json.Marshal(essentialInfo)
			} else {
				essentialInfo := extractEssentialInfoList(things.Items)
				for i := range essentialInfo {
					markEmpty(&essentialInfo[i])
				}
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
