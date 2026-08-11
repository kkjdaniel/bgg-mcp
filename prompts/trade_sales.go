package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterPrompts(s *server.MCPServer) {
	tradeSalesPrompt := mcp.NewPrompt("trade-sales-post",
		mcp.WithPromptDescription("Generate a sales post for your BGG 'for trade' collection with discounted prices"),
		mcp.WithArgument("username",
			mcp.ArgumentDescription("BoardGameGeek username"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("currency",
			mcp.ArgumentDescription("Currency for prices (USD, GBP, EUR) - default: USD"),
		),
		mcp.WithArgument("destination",
			mcp.ArgumentDescription("Destination country (US, GB, DE) - default: US"),
		),
	)

	tradeSalesHandler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		username := request.Params.Arguments["username"]
		if username == "" {
			return nil, fmt.Errorf("username is required")
		}

		currency := request.Params.Arguments["currency"]
		if currency == "" {
			currency = "USD"
		}

		destination := request.Params.Arguments["destination"]
		if destination == "" {
			destination = "US"
		}

		return mcp.NewGetPromptResult(
			"Generate BGG trade collection sales post",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(fmt.Sprintf(`Please help me create a sales post for my BoardGameGeek for-trade collection. Here's what I need:

1. First, use the bgg-collection tool to fetch my collection with username "%s" and filter for games marked "fortrade"
2. For each game in the collection, use the bgg-price tool to get current prices in %s currency for %s destination
3. Create a formatted sales post with:
   - A header saying "🎲 BOARD GAMES FOR SALE 🎲"
   - List each game with its name and price (reduce prices by 20%% for a quick sale)
   - Use emoji status indicators: 🟢 = Available, 🟡 = Pending, 🔴 = Sold (default all to 🟢)
   - If price is not available, show "Price TBD"
   - End with "DM for more info or bundle deals!"

Format it nicely for posting in a Facebook hobby group.`, username, currency, destination)),
				),
			},
		), nil
	}

	s.AddPrompt(tradeSalesPrompt, tradeSalesHandler)

	RegisterGameRecommendationsPrompt(s)
}