package prompts

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	usernameArgumentDescription    = "BoardGameGeek username (defaults to the configured BGG_USERNAME)"
	currencyArgumentDescription    = "Currency for prices: USD, GBP, EUR, DKK or SEK (default: USD)"
	destinationArgumentDescription = "Destination country for prices: US, GB, DE, DK or SE (default: US)"
)

// RegisterPrompts registers every prompt. defaultUsername is used by prompts
// when no username argument is given.
func RegisterPrompts(s *server.MCPServer, defaultUsername string) {
	RegisterTradeSalesPrompt(s, defaultUsername)
	RegisterGameRecommendationsPrompt(s, defaultUsername)
	RegisterRulesQuestionPrompt(s)
}

func argument(request mcp.GetPromptRequest, name string, fallback string) string {
	if value := strings.TrimSpace(request.Params.Arguments[name]); value != "" {
		return value
	}
	return fallback
}

func usernameArgument(request mcp.GetPromptRequest, defaultUsername string) (string, error) {
	username := argument(request, "username", defaultUsername)
	if username == "" {
		return "", fmt.Errorf("username is required when BGG_USERNAME is not set")
	}
	return username, nil
}
