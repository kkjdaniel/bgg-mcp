package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func PriceTool() (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-price",
		mcp.WithDescription("Get current prices for board games from multiple retailers using BGG IDs"),
		mcp.WithString("ids",
			mcp.Required(),
			mcp.Description("Comma-separated BGG IDs (e.g., '12,844,2096,13857')"),
		),
		mcp.WithString("currency",
			mcp.Description("Currency code (default: USD)"),
			mcp.Enum("DKK", "GBP", "SEK", "EUR", "USD"),
		),
		mcp.WithString("destination",
			mcp.Description("Destination country (default: US)"),
			mcp.Enum("DK", "SE", "GB", "DE", "US"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()

		ids, hasIDs := arguments["ids"].(string)
		if !hasIDs || ids == "" {
			return mcp.NewToolResultText("IDs parameter is required"), nil
		}

		currency := "USD"
		if c, ok := arguments["currency"].(string); ok {
			currency = strings.ToUpper(c)
		}

		destination := "US"
		if d, ok := arguments["destination"].(string); ok {
			destination = strings.ToUpper(d)
		}

		baseURL := "https://boardgameprices.co.uk/api/info"
		params := url.Values{}
		params.Add("eid", ids)
		params.Add("currency", currency)
		params.Add("destination", destination)
		params.Add("sitename", "bgg-mcp")

		resp, err := http.Get(baseURL + "?" + params.Encode())
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("API request error: %v", err)), nil
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error reading response: %v", err)), nil
		}

		var result interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("JSON parsing error: %v", err)), nil
		}

		out, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("JSON encoding error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}

	return tool, handler
}
