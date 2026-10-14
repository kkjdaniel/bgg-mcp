package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/collection"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CollectionResult is the matching games in a user's collection, with links to
// the user's BGG profile, their collection and each game.
type CollectionResult struct {
	Username       string           `json:"username"`
	UserLink       string           `json:"user_link"`
	CollectionLink string           `json:"collection_link"`
	Count          int              `json:"count"`
	Items          []CollectionGame `json:"items"`
}

// CollectionGame is a game in a user's collection: the user's own rating, plays
// and status for it, BGG's figures for the game, and the link to its BGG page.
// It leaves out what the collection endpoint returns that answers rarely use,
// such as image URLs and every family rank, since a collection can run to
// hundreds of games.
type CollectionGame struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	Link           string   `json:"link"`
	Year           int      `json:"year,omitempty"`
	PersonalRating float64  `json:"personal_rating,omitempty"`
	BGGRating      float64  `json:"bgg_rating,omitempty"`
	BGGRank        int      `json:"bgg_rank,omitempty"`
	Players        string   `json:"players,omitempty"`
	PlayTime       string   `json:"play_time,omitempty"`
	NumPlays       int      `json:"num_plays"`
	Status         []string `json:"status"`
	Comment        string   `json:"comment,omitempty"`
}

func NewCollectionResult(username string, items []collection.CollectionItem) CollectionResult {
	games := make([]CollectionGame, 0, len(items))
	for _, item := range items {
		games = append(games, newCollectionGame(item))
	}

	return CollectionResult{
		Username:       username,
		UserLink:       profileURL(username),
		CollectionLink: collectionURL(username),
		Count:          len(games),
		Items:          games,
	}
}

func newCollectionGame(item collection.CollectionItem) CollectionGame {
	game := CollectionGame{
		ID:       item.ObjectID,
		Name:     item.Name,
		Link:     itemURL(item.Subtype, item.ObjectID),
		Year:     item.YearPublished,
		NumPlays: item.NumPlays,
		Status:   collectionStatus(item.Status),
		Comment:  item.Comment,
	}

	// Stats only come back from a query made WithStats.
	if stats := item.Stats; stats != nil {
		// An unrated game's rating is "N/A", which leaves the field out.
		game.PersonalRating = ownRating(item)
		game.BGGRating = bggRating(item)
		for _, rank := range stats.Rating.Ranks {
			// The overall rank; the rest are family ranks such as strategy games.
			// An unranked game's value is "Not Ranked", which leaves it out.
			if rank.Name == "boardgame" {
				game.BGGRank, _ = strconv.Atoi(rank.Value)
			}
		}
		game.Players = playerRange(stats.MinPlayers, stats.MaxPlayers)
		game.PlayTime = playTimeRange(stats.MinPlayTime, stats.MaxPlayTime)
	}

	return game
}

// collectionStatus names the status flags set on a collection item.
func collectionStatus(status collection.ItemStatus) []string {
	flags := []struct {
		set   int
		label string
	}{
		{status.Own, "owned"},
		{status.PrevOwned, "previously owned"},
		{status.Preordered, "preordered"},
		{status.ForTrade, "for trade"},
		{status.Want, "wanted in trade"},
		{status.WantToPlay, "want to play"},
		{status.WantToBuy, "want to buy"},
		{status.Wishlist, "wishlist"},
	}

	labels := []string{}
	for _, flag := range flags {
		if flag.set == 1 {
			labels = append(labels, flag.label)
		}
	}
	return labels
}

func CollectionTool(client *gogeek.Client, selfUsername string) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-collection",
		mcp.WithDescription(withLinkNote("Query a user's board game collection on BoardGameGeek (BGG). Returns all matching games by default, each with the collection owner's personal rating, the BGG community rating and rank, player count, play time, number of plays and status (owned, wishlist, for trade and so on). Use the filter parameters to narrow results (e.g. owned, wishlist, rated, play count). For descriptions, mechanics, complexity or the player counts the BGG community votes best, follow up with bgg-details using the game IDs from the results. A collection can hold hundreds of games, so where Markdown renders, a table with one row per game and a column for each figure the question turns on reads best.")),
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
			if selfUsername == "" {
				return mcp.NewToolResultText(selfUsernameMissingMessage), nil
			}
			username = selfUsername
		}

		options := buildCollectionOptions(arguments)

		result, err := collection.Query(ctx, client, username, options...)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error fetching collection: %v", err)), nil
		}

		if len(result.Items) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No items found in collection with the specified filters. Collection on BGG: %s", collectionURL(username))), nil
		}

		out, err := json.Marshal(NewCollectionResult(username, result.Items))
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}

func buildCollectionOptions(arguments map[string]interface{}) []collection.Option {
	// Stats carry the personal and BGG ratings, rank, player count and play time.
	options := []collection.Option{collection.WithStats()}

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

	booleanFilters := map[string]func(bool) collection.Option{
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

	numericFilters := map[string]func(float64) collection.Option{
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
