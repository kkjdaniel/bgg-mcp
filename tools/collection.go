package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kkjdaniel/gogeek/v2"
	"github.com/kkjdaniel/gogeek/v2/collection"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func CollectionTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-collection",
		mcp.WithDescription("Query a user's board game collection on BoardGameGeek (BGG). Returns all matching games by default. Use the filter parameters to narrow results (e.g. owned, wishlist, rated, play count)."),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the BoardGameGeek (BGG) user who owns the collection. When the user refers to themselves (me, my, I), use 'SELF' as the value."),
		),
		mcp.WithString("subtype",
			mcp.Enum("boardgame", "boardgameexpansion"),
			mcp.Description("Filter by game type: 'boardgame' for base games only (excludes expansions), 'boardgameexpansion' for expansions only"),
		),
		mcp.WithBoolean("owned",
			mcp.Description("Filters for owned games in the collection (default: true if no ownership filters specified)"),
		),
		mcp.WithBoolean("wishlist",
			mcp.Description("Filters for wishlisted games in the collection"),
		),
		mcp.WithBoolean("preordered",
			mcp.Description("Filters for preordered games in the collection"),
		),
		mcp.WithBoolean("fortrade",
			mcp.Description("Filters for games that are marked for trade in the collection"),
		),
		mcp.WithBoolean("rated",
			mcp.Description("Filters for games that are rated in the collection"),
		),
		mcp.WithBoolean("wanttoplay",
			mcp.Description("Filters for games that the user wants to play in the collection"),
		),
		mcp.WithBoolean("played",
			mcp.Description("Filters for games that have recorded plays in the collection"),
		),
		mcp.WithBoolean("wanttobuy",
			mcp.Description("Filters for games that the user wants to buy in the collection"),
		),
		mcp.WithBoolean("hasparts",
			mcp.Description("Filters for games that have spare parts or not in the collection"),
		),
		mcp.WithNumber("minrating",
			mcp.Description("Filters based on the minimum personal rating of the games in the collection"),
		),
		mcp.WithNumber("maxrating",
			mcp.Description("Filters based on the maximum personal rating of the games in the collection"),
		),
		mcp.WithNumber("minbggrating",
			mcp.Description("Filters based on the minimum global BoardGameGeek (BGG) rating of the games in the collection"),
		),
		mcp.WithNumber("maxbggrating",
			mcp.Description("Filters based on the maximum global BoardGameGeek (BGG) rating of the games in the collection"),
		),
		mcp.WithNumber("minplays",
			mcp.Description("Filters based on the minimum number of plays of the games in the collection"),
		),
		mcp.WithNumber("maxplays",
			mcp.Description("Filters based on the maximum number of plays of the games in the collection"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()

		username, ok := arguments["username"].(string)
		if !ok || username == "" {
			return mcp.NewToolResultText("Username is required"), nil
		}

		if username == "SELF" {
			envUsername := os.Getenv("BGG_USERNAME")
			if envUsername == "" {
				return mcp.NewToolResultText("BGG_USERNAME environment variable not set. Either set it or provide your specific username instead of 'SELF'."), nil
			}
			username = envUsername
		}

		options := buildCollectionOptions(arguments)

		result, err := collection.Query(client, username, options...)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error fetching collection: %v", err)), nil
		}

		if len(result.Items) == 0 {
			return mcp.NewToolResultText("No items found in collection with the specified filters"), nil
		}

		out, err := json.Marshal(result.Items)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}

func buildCollectionOptions(arguments map[string]interface{}) []collection.CollectionOption {
	var options []collection.CollectionOption

	ownershipFilters := []string{"owned", "wishlist", "preordered", "fortrade", "wanttoplay", "wanttobuy"}
	hasOwnershipFilter := false
	for _, filter := range ownershipFilters {
		if arguments[filter] != nil {
			hasOwnershipFilter = true
			break
		}
	}

	if !hasOwnershipFilter {
		options = append(options, collection.WithOwned(true))
	}

	if subtype, ok := arguments["subtype"].(string); ok {
		if subtype == "boardgame" {
			options = append(options, collection.WithSubtype("boardgame"))
			options = append(options, collection.WithExcludeSubtype("boardgameexpansion"))
		} else {
			options = append(options, collection.WithSubtype(subtype))
		}
	}

	booleanFilters := map[string]func(bool) collection.CollectionOption{
		"owned":      collection.WithOwned,
		"wishlist":   collection.WithWishlist,
		"preordered": collection.WithPreordered,
		"fortrade":   collection.WithTrade,
		"rated":      collection.WithRated,
		"wanttoplay": collection.WithWantToPlay,
		"played":     collection.WithPlayed,
		"wanttobuy":  collection.WithWantToBuy,
		"hasparts":   collection.WithHasParts,
	}

	for key, fn := range booleanFilters {
		if val, ok := arguments[key].(bool); ok {
			options = append(options, fn(val))
		}
	}

	numericFilters := map[string]func(float64) collection.CollectionOption{
		"minrating":    collection.WithMinRating,
		"maxrating":    collection.WithMaxRating,
		"minbggrating": collection.WithMinBGGRating,
		"maxbggrating": collection.WithMaxBGGRating,
	}

	for key, fn := range numericFilters {
		if val, ok := arguments[key].(float64); ok {
			options = append(options, fn(val))
		}
	}

	if minplays, ok := arguments["minplays"].(float64); ok {
		options = append(options, collection.WithMinPlays(int(minplays)))
	}

	if maxplays, ok := arguments["maxplays"].(float64); ok {
		options = append(options, collection.WithMaxPlays(int(maxplays)))
	}

	return options
}
