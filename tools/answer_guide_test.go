package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestAnswerGuideMiddleware(t *testing.T) {
	respond := func(text string) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(text), nil
		}
	}

	cases := []struct {
		name      string
		text      string
		wantGuide bool
	}{
		{"result with BGG links", `{"name":"Catan","link":"https://boardgamegeek.com/boardgame/13"}`, true},
		{"prices from another site", `{"price":45,"link":"https://boardgameprices.co.uk/item/13"}`, false},
		{"error without links", "Username is required", false},
	}
	for _, c := range cases {
		result, err := AnswerGuideMiddleware(respond(c.text))(context.Background(), mcp.CallToolRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if got := len(result.Content) == 2; got != c.wantGuide {
			t.Errorf("%s: guide added = %v, want %v", c.name, got, c.wantGuide)
		}
		if first := result.Content[0].(mcp.TextContent).Text; first != c.text {
			t.Errorf("%s: the result itself should be unchanged, got %q", c.name, first)
		}
		if c.wantGuide {
			guide := result.Content[1].(mcp.TextContent)
			if guide.Annotations == nil || len(guide.Annotations.Audience) != 1 || guide.Annotations.Audience[0] != mcp.RoleAssistant {
				t.Errorf("%s: the guide should be marked for the assistant only, got %+v", c.name, guide.Annotations)
			}
		}
	}
}
