package prompts

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterRulesQuestionPrompt(s *server.MCPServer) {
	rulesQuestionPrompt := mcp.NewPrompt("rules-question",
		mcp.WithPromptDescription("Answer a board game rules question using the game's BGG rules forum"),
		mcp.WithArgument("game",
			mcp.ArgumentDescription("The name of the board game"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("question",
			mcp.ArgumentDescription("The rules question to answer"),
			mcp.RequiredArgument(),
		),
	)

	rulesQuestionHandler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		game := argument(request, "game", "")
		if game == "" {
			return nil, fmt.Errorf("game is required")
		}

		question := argument(request, "question", "")
		if question == "" {
			return nil, fmt.Errorf("question is required")
		}

		return mcp.NewGetPromptResult(
			"Answer a board game rules question",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(fmt.Sprintf(`I have a rules question about the board game "%s":

%s

Please answer it from the game's BoardGameGeek rules forum rather than from memory:

1. Use the bgg-rules tool with name "%s" and the question above as the 'question' parameter.
2. Check that the game it found is the one I mean. If it is not, tell me and try again with a more specific name.
3. Read the posts included with the top matches. If they do not settle the question, use the bgg-thread-details tool on the next most relevant thread or on a reference thread (FAQ or errata).
4. Give me a direct answer first, then a short explanation of the reasoning.
5. Link the thread or threads the answer comes from, and say whether it comes from the designer or publisher, from the rulebook as quoted by posters, or from community consensus.

If the forum does not answer the question, say so plainly, and only then give your own best reading of the rules, clearly marked as such.`, game, question, game)),
				),
			},
		), nil
	}

	s.AddPrompt(rulesQuestionPrompt, rulesQuestionHandler)
}
