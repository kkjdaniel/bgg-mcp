package prompts

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterTradeSalesPrompt(s *server.MCPServer, defaultUsername string) {
	tradeSalesPrompt := mcp.NewPrompt("trade-sales-post",
		mcp.WithPromptDescription("Generate a sales post for your BGG 'for trade' collection, priced at a discount to current retail prices"),
		mcp.WithArgument("username",
			mcp.ArgumentDescription(usernameArgumentDescription),
		),
		mcp.WithArgument("currency",
			mcp.ArgumentDescription(currencyArgumentDescription),
		),
		mcp.WithArgument("destination",
			mcp.ArgumentDescription(destinationArgumentDescription),
		),
		mcp.WithArgument("discount",
			mcp.ArgumentDescription("Percentage to take off the current retail price (default: 20)"),
		),
		mcp.WithArgument("platform",
			mcp.ArgumentDescription("Where the post will be published, e.g. 'BoardGameGeek marketplace' or 'Facebook group' (default: BoardGameGeek marketplace)"),
		),
	)

	tradeSalesHandler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		username, err := usernameArgument(request, defaultUsername)
		if err != nil {
			return nil, err
		}

		currency := argument(request, "currency", "USD")
		destination := argument(request, "destination", "US")
		platform := argument(request, "platform", "BoardGameGeek marketplace")

		discount, err := strconv.ParseFloat(strings.TrimSuffix(argument(request, "discount", "20"), "%"), 64)
		if err != nil || discount < 0 || discount >= 100 {
			return nil, fmt.Errorf("discount must be a percentage between 0 and 100")
		}

		return mcp.NewGetPromptResult(
			"Generate BGG trade collection sales post",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(fmt.Sprintf(`Please help me create a sales post for my BoardGameGeek for-trade collection. Here's what I need:

1. Use the bgg-collection tool with username "%s" and fortrade set to true to fetch the games I have marked for trade.
2. Use the bgg-collection tool again with the same username, fortrade set to true and subtype "boardgameexpansion" to find out which of those are expansions.
3. Use the bgg-price tool once, passing all of the game and expansion IDs as a single comma-separated list, to get current retail prices in %s currency for %s destination.
4. Sanity-check every price before using it. Treat a price as suspect when it is far out of line with what a game of that size and age normally sells for, when it comes from a single listing with nothing to compare against, or when one listing is several times higher or lower than the others. Where other listings exist, ignore the outlier and use the lowest sensible price. Where the only price is suspect, do not build an asking price from it: show "Price TBD" instead.
5. Bundle expansions with their base game. When an expansion and the game it expands are both for trade, list them as one bundle with a single asking price based on the combined retail prices, and name the expansions included. List an expansion on its own only when its base game is not for trade, and mark it as an expansion that needs the base game.
6. Create a sales post with:
   - A header saying "🎲 BOARD GAMES FOR SALE 🎲"
   - Each game or bundle with its name and asking price, where the asking price is the retail price (combined, for a bundle) reduced by %s%% and rounded to a sensible whole number
   - Emoji status indicators: 🟢 = Available, 🟡 = Pending, 🔴 = Sold (default all to 🟢)
   - "Price TBD" for any game with no usable retail price
   - A closing line saying "DM for more info or bundle deals!"

Format it for posting on: %s. Use the formatting conventions that work there.

After the post, add a short "Prices to check" note for me that is not part of the post. In it, list every price you treated as suspect, with the figure you saw and why it looked wrong (for example "only one listing, at £500"), any bundle where one of the items had no price, and any other game where a new-copy retail price is a poor guide, such as out-of-print games. If there is nothing to flag, say so.`,
						username, currency, destination, strconv.FormatFloat(discount, 'f', -1, 64), platform)),
				),
			},
		), nil
	}

	s.AddPrompt(tradeSalesPrompt, tradeSalesHandler)
}
