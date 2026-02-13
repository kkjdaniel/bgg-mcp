package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kkjdaniel/gogeek/v2"
	"github.com/kkjdaniel/gogeek/v2/search"
	"github.com/kkjdaniel/gogeek/v2/thing"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func SearchTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-search",
		mcp.WithDescription("Search for board games on BoardGameGeek (BGG) by name or part of a name using a broad search (e.g., 'Catan', 'Ticket to Ride')"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Game name to search for on BoardGameGeek (BGG)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return (default: 30)"),
		),
		mcp.WithString("type",
			mcp.Description("Filter results by game type (default: all)"),
			mcp.Enum("all", "boardgame", "boardgameexpansion"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		query := arguments["query"].(string)

		limit := 30
		if l, ok := arguments["limit"].(float64); ok {
			limit = int(l)
		}

		typeFilter := "all"
		if t, ok := arguments["type"].(string); ok {
			typeFilter = t
		}

		gameDetails, err := searchAndSortGames(client, query, typeFilter, limit)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		essentialInfo := extractEssentialInfoList(gameDetails.Items)
		out, err := json.Marshal(essentialInfo)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("JSON encoding error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}

func searchAndSortGames(client *gogeek.Client, query, typeFilter string, limit int) (*thing.Items, error) {
	result, err := search.Query(client, query, false)
	if err != nil {
		return nil, fmt.Errorf("search error: %v", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no search results found")
	}

	var filteredItems []search.SearchResult
	if typeFilter == "all" {
		filteredItems = result.Items
	} else {
		for _, item := range result.Items {
			if item.Type == typeFilter {
				filteredItems = append(filteredItems, item)
			}
		}
	}

	if len(filteredItems) == 0 {
		return nil, fmt.Errorf("no %s results found", typeFilter)
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))

	var exactMatches []search.SearchResult
	for _, item := range filteredItems {
		if strings.ToLower(item.Name.Value) == queryLower {
			exactMatches = append(exactMatches, item)
		}
	}

	if len(exactMatches) > 0 {
		filteredItems = exactMatches
	} else {
		var baseGames, expansions []search.SearchResult
		delimiters := []string{":", " – ", " - ", " — ", " (", " ["}

		for _, item := range filteredItems {
			nameLower := strings.ToLower(item.Name.Value)
			firstPart := nameLower

			for _, delimiter := range delimiters {
				if index := strings.Index(nameLower, delimiter); index > 0 {
					if candidate := strings.TrimSpace(nameLower[:index]); len(candidate) < len(firstPart) {
						firstPart = candidate
					}
				}
			}

			if firstPart == queryLower && item.Type == "boardgame" {
				baseGames = append(baseGames, item)
			} else {
				expansions = append(expansions, item)
			}
		}

		filteredItems = append(baseGames, expansions...)
	}

	gameIDs := make([]int, 0, len(filteredItems))
	for _, item := range filteredItems {
		gameIDs = append(gameIDs, item.ID)
	}

	var allItems []thing.Item
	maxBatch := 20

	for i := 0; i < len(gameIDs); i += maxBatch {
		end := i + maxBatch
		if end > len(gameIDs) {
			end = len(gameIDs)
		}

		batch := gameIDs[i:end]
		gameDetails, err := thing.Query(client, batch)
		if err != nil {
			return nil, fmt.Errorf("error fetching game details: %v", err)
		}

		allItems = append(allItems, gameDetails.Items...)
	}

	gameDetails := &thing.Items{Items: allItems}

	sort.Slice(gameDetails.Items, func(i, j int) bool {
		votesI := 0
		votesJ := 0

		if gameDetails.Items[i].Statistics != nil {
			votesI = gameDetails.Items[i].Statistics.UsersRated.Value
		}
		if gameDetails.Items[j].Statistics != nil {
			votesJ = gameDetails.Items[j].Statistics.UsersRated.Value
		}

		return votesI > votesJ
	})

	if len(gameDetails.Items) > limit {
		gameDetails.Items = gameDetails.Items[:limit]
	}

	return gameDetails, nil
}
