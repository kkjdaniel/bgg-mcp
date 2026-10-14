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

type TradeOpportunity struct {
	User1Username    string          `json:"user1_username"`
	User1Link        string          `json:"user1_link"`
	User2Username    string          `json:"user2_username"`
	User2Link        string          `json:"user2_link"`
	User1HasWanted   []TradeItem     `json:"user1_has_wanted"`
	User2Wishlist    []TradeItem     `json:"user2_wishlist"`
	Summary          TradeSummary    `json:"summary"`
}

type TradeItem struct {
	GameID       int     `json:"game_id"`
	Name         string  `json:"name"`
	Link         string  `json:"link"`
	YearPublished int    `json:"year_published"`
	ForTrade     bool    `json:"for_trade"`
	WantInTrade  bool    `json:"want_in_trade"`
	UserRating   float64 `json:"user_rating,omitempty"`
	BGGRating    float64 `json:"bgg_rating,omitempty"`
}

type BasicGameInfo struct {
	GameID       int    `json:"game_id"`
	Name         string `json:"name"`
	YearPublished int   `json:"year_published"`
}

type TradeSummary struct {
	User1HasWantedCount int  `json:"user1_has_wanted_count"`
	User2WishlistCount  int  `json:"user2_wishlist_count"`
	HasTradeOpportunity bool `json:"has_trade_opportunity"`
}

func TradeFinderTool(client *gogeek.Client, selfUsername string) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-trade-finder",
		mcp.WithDescription(withLinkNote("Find what games user1 owns that user2 has on their wishlist. Shows potential trading opportunities, with each game's BGG rating and the owner's own rating ('user_rating': user1's for games they own, user2's for their wishlist).")),
		mcp.WithString("user1",
			mcp.Required(),
			mcp.Description("BGG username whose collection will be checked. When the user refers to themselves (me, my, I), use 'SELF' as the value."),
		),
		mcp.WithString("user2", 
			mcp.Required(),
			mcp.Description("BGG username whose wishlist will be checked against user1's collection"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()

		user1, ok := arguments["user1"].(string)
		if !ok || user1 == "" {
			return mcp.NewToolResultText("user1 is required"), nil
		}

		if user1 == "SELF" {
			if selfUsername == "" {
				return mcp.NewToolResultText(selfUsernameMissingMessage), nil
			}
			user1 = selfUsername
		}

		user2, ok := arguments["user2"].(string)
		if !ok || user2 == "" {
			return mcp.NewToolResultText("user2 is required"), nil
		}

		if user2 == "SELF" {
			if selfUsername == "" {
				return mcp.NewToolResultText(selfUsernameMissingMessage), nil
			}
			user2 = selfUsername
		}

		// Stats carry each user's own rating and the BGG rating of every game.
		user1Collection, err := collection.Query(ctx, client, user1, collection.WithOwned(true), collection.WithStats())
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error fetching %s's collection: %v", user1, err)), nil
		}

		user2Wishlist, err := collection.Query(ctx, client, user2, collection.WithWishlist(true), collection.WithStats())
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error fetching %s's wishlist: %v", user2, err)), nil
		}

		tradeAnalysis := analyseTradeOpportunities(user1, user2, user1Collection, user2Wishlist)

		out, err := json.Marshal(tradeAnalysis)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}

func analyseTradeOpportunities(user1, user2 string, user1Col, user2WishlistCol *collection.Collection) TradeOpportunity {
	user2WishlistMap := make(map[int]*collection.CollectionItem)
	for _, item := range user2WishlistCol.Items {
		user2WishlistMap[item.ObjectID] = &item
	}

	var user1HasWanted []TradeItem
	var user2Wishlist []TradeItem

	for _, user1Item := range user1Col.Items {
		if _, exists := user2WishlistMap[user1Item.ObjectID]; exists {
			user1HasWanted = append(user1HasWanted, TradeItem{
				GameID:       user1Item.ObjectID,
				Name:         user1Item.Name,
				Link:         itemURL(user1Item.Subtype, user1Item.ObjectID),
				YearPublished: user1Item.YearPublished,
				ForTrade:     user1Item.Status.ForTrade == 1,
				WantInTrade:  user1Item.Status.Want == 1,
				UserRating:   ownRating(user1Item),
				BGGRating:    bggRating(user1Item),
			})
		}
	}

	for _, wishlistItem := range user2WishlistCol.Items {
		user2Wishlist = append(user2Wishlist, TradeItem{
			GameID:       wishlistItem.ObjectID,
			Name:         wishlistItem.Name,
			Link:         itemURL(wishlistItem.Subtype, wishlistItem.ObjectID),
			YearPublished: wishlistItem.YearPublished,
			ForTrade:     wishlistItem.Status.ForTrade == 1,
			WantInTrade:  true,
			UserRating:   ownRating(wishlistItem),
			BGGRating:    bggRating(wishlistItem),
		})
	}

	if len(user1HasWanted) == 0 {
		return TradeOpportunity{
			User1Username: user1,
			User1Link: profileURL(user1),
			User2Username: user2,
			User2Link: profileURL(user2),
			User1HasWanted: []TradeItem{},
			User2Wishlist: user2Wishlist,
			Summary: TradeSummary{
				User1HasWantedCount: 0,
				User2WishlistCount: len(user2Wishlist),
				HasTradeOpportunity: false,
			},
		}
	}

	summary := TradeSummary{
		User1HasWantedCount: len(user1HasWanted),
		User2WishlistCount: len(user2Wishlist),
		HasTradeOpportunity: true,
	}

	return TradeOpportunity{
		User1Username: user1,
		User1Link: profileURL(user1),
		User2Username: user2,
		User2Link: profileURL(user2),
		User1HasWanted: user1HasWanted,
		User2Wishlist: user2Wishlist,
		Summary:        summary,
	}
}

// ownRating is the rating the collection's owner gave a game, or 0 when they
// haven't rated it.
func ownRating(item collection.CollectionItem) float64 {
	if item.Stats == nil {
		return 0
	}
	rating, _ := strconv.ParseFloat(item.Stats.Rating.Value, 64)
	return rating
}

// bggRating is a game's average rating on BGG, or 0 when it has none.
func bggRating(item collection.CollectionItem) float64 {
	if item.Stats == nil {
		return 0
	}
	rating, _ := strconv.ParseFloat(item.Stats.Rating.Average.Value, 64)
	return rating
}
