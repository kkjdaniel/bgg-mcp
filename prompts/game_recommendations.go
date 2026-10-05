package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterGameRecommendationsPrompt(s *server.MCPServer, defaultUsername string) {
	gameRecommendationPrompt := mcp.NewPrompt("game-recommendations",
		mcp.WithPromptDescription("Get personalized board game recommendations based on your BGG collection and preferences"),
		mcp.WithArgument("username",
			mcp.ArgumentDescription(usernameArgumentDescription),
		),
		mcp.WithArgument("currency",
			mcp.ArgumentDescription(currencyArgumentDescription),
		),
		mcp.WithArgument("destination",
			mcp.ArgumentDescription(destinationArgumentDescription),
		),
	)

	gameRecommendationHandler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		username, err := usernameArgument(request, defaultUsername)
		if err != nil {
			return nil, err
		}

		currency := argument(request, "currency", "USD")
		destination := argument(request, "destination", "US")

		return mcp.NewGetPromptResult(
			"Board Game Recommendation Expert",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(fmt.Sprintf(`You are a board game recommendation expert. Please provide personalized game recommendations for BoardGameGeek user "%s" by following these steps:

1. **Get what they own**: Use the bgg-collection tool with username "%s" and owned set to true. Keep this list: nothing on it may be recommended.

2. **Find their favourites**: Use the bgg-collection tool with username "%s", rated set to true and minrating 8, then take up to five of the highest-rated games. If that returns nothing, lower minrating to 7. If they have not rated any games, use their most played owned games instead (minplays 1), and if there is still nothing to go on, ask them to name a few games they love before continuing.

3. **Generate recommendations**: For each favourite, use the bgg-recommender tool with the game's ID to get similar games.

4. **Curate the list**: Remove every game they already own, then choose five to eight recommendations, taking one or two per favourite so the list covers a variety of genres and mechanics.

5. **Get the details**: Use the bgg-details tool once, passing all the chosen game IDs in the 'ids' parameter, to get each game's year, mechanics and complexity.

6. **Get pricing**: Use the bgg-price tool once, passing all the chosen game IDs as a single comma-separated list, to get current prices in %s currency for %s destination.

7. **Format the response** as shown below:

## 🎲 Your Game Recommendations

Based on your love of [list 2-3 of their favourite games], here are my recommendations:

### 1. **Brass: Birmingham** (2018)
*Perfect for fans of economic strategy - build industries and rail networks in Industrial Revolution England*
- **Recommended because**: you rated [favourite game] highly
- **Mechanisms**: Network building, Hand management, Economic
- **Complexity**: 3.9/5
- **Best Price**: [$67.99](link)

[Continue for each recommendation...]

**Guidelines:**
- Keep descriptions compelling and focused on why THEY would enjoy it
- Take mechanisms and complexity from the bgg-details results, not from memory
- Show "Price not found" when bgg-price has no price for a game
- Never recommend a game from their owned list`, username, username, username, currency, destination)),
				),
			},
		), nil
	}

	s.AddPrompt(gameRecommendationPrompt, gameRecommendationHandler)
}
