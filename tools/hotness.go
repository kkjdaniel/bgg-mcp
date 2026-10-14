package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/hot"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HotGame is one game on the BGG hotness list.
type HotGame struct {
	Rank      int    `json:"rank"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Link      string `json:"link"`
	Year      int    `json:"year,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

// FetchHotGames returns the current BGG hotness list for board games.
func FetchHotGames(ctx context.Context, client *gogeek.Client) ([]HotGame, error) {
	hotItems, err := hot.Query(ctx, client, hot.ItemTypeBoardGame)
	if err != nil {
		return nil, err
	}

	games := make([]HotGame, 0, len(hotItems.Items))
	for _, item := range hotItems.Items {
		games = append(games, HotGame{
			Rank:      item.Rank,
			ID:        item.ID,
			Name:      item.Name.Value,
			Link:      itemURL(string(hot.ItemTypeBoardGame), item.ID),
			Year:      item.YearPublished.Value,
			Thumbnail: item.Thumbnail.Value,
		})
	}
	return games, nil
}

func HotnessTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-hot",
		mcp.WithDescription("Find the current trending board game hotness on BoardGameGeek (BGG). Returns each game's hotness rank, name, year and link only; for ratings, player counts, complexity or what a game is about, follow up with bgg-details using 'ids' (up to 20 at a time)."),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		hotGames, err := FetchHotGames(ctx, client)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		if len(hotGames) > 0 {
			out, err := json.Marshal(hotGames)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
			}
			return mcp.NewToolResultText(string(out)), nil
		}

		return mcp.NewToolResultText("No hot games found"), nil
	}

	return tool, handler
}
