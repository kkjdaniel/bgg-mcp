package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kkjdaniel/gogeek/v2"
	"github.com/kkjdaniel/gogeek/v2/thing"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type RecommendGamesResponse struct {
	Results []RecommendGameItem `json:"results"`
}

type RecommendGameItem struct {
	BGGID int `json:"bgg_id"`
}

func RecommenderTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-recommender",
		mcp.WithDescription("Get game recommendations similar to a specific board game. Provide either 'name' or 'id', not both."),
		mcp.WithString("name",
			mcp.Description("Name of the game to base recommendations on. Use when the BGG ID is not known."),
		),
		mcp.WithNumber("id",
			mcp.Description("BoardGameGeek (BGG) ID of the game to base recommendations on. Preferred over 'name' when already known."),
		),
		mcp.WithNumber("min_votes",
			mcp.Description("Minimum votes threshold for recommendation quality (default: 30)"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()

		var gameID int
		var err error

		if nameVal, ok := arguments["name"].(string); ok && nameVal != "" {
			gameDetails, err := searchAndSortGames(client, nameVal, "boardgame", 1)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error finding game by name: %v", err)), nil
			}
			if len(gameDetails.Items) == 0 {
				return mcp.NewToolResultText("No games found with that name"), nil
			}
			gameID = gameDetails.Items[0].ID
		} else if idVal, ok := arguments["id"].(float64); ok && idVal > 0 {
			gameID = int(idVal)
		} else {
			return mcp.NewToolResultText("Either 'name' or 'id' parameter must be provided"), nil
		}

		minVotes := 30
		if mv, ok := arguments["min_votes"].(float64); ok && mv > 0 {
			minVotes = int(mv)
		}

		recommendURL := fmt.Sprintf("https://recommend.games/api/games/%d/similar.json?num_votes__gte=%d&page=1", gameID, minVotes)

		resp, err := http.Get(recommendURL)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error fetching recommendations: %v", err)), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return mcp.NewToolResultText(fmt.Sprintf("Recommendation API returned status %d", resp.StatusCode)), nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error reading recommendation response: %v", err)), nil
		}

		var recResponse RecommendGamesResponse
		if err := json.Unmarshal(body, &recResponse); err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error parsing recommendation response: %v", err)), nil
		}

		recommendedIDs := make([]int, 0, 10)
		for i, game := range recResponse.Results {
			if i >= 10 {
				break
			}
			recommendedIDs = append(recommendedIDs, game.BGGID)
		}

		if len(recommendedIDs) == 0 {
			return mcp.NewToolResultText("No recommendations found"), nil
		}

		gameDetails, err := thing.Query(client, recommendedIDs)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error fetching game details: %v", err)), nil
		}

		essentialInfo := extractEssentialInfoList(gameDetails.Items)
		out, err := json.Marshal(essentialInfo)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}
