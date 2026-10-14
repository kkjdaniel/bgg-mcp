package tools

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// answerGuide repeats the heart of the server instructions after each result.
// Not every client hands the model the server instructions, and answers came
// back with no links to BGG and long lists as bullets, but every client passes
// tool results on, so this is the one place the model is sure to see it.
// It asks for Markdown only where the interface renders it, since some clients
// show plain text.
const answerGuide = `Note for the assistant on presenting this result (don't show or quote it to the user): BoardGameGeek asks that answers built on its data link back to it, so give every game, designer, publisher, user or thread you name its "link" from the result above, never the name alone. Where your interface renders Markdown, put the link on the name, such as [Brass: Birmingham](https://boardgamegeek.com/boardgame/224517), and lay out more than a handful of games as a table, one row per game, with a column for each figure the question turns on. Where it shows plain text, put the URL after the name instead, and list one game per line with its figures.`

// AnswerGuideMiddleware adds answerGuide after any tool result that carries
// links to BGG. Results without them, such as prices from another site or an
// error, are left alone.
func AnswerGuideMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := next(ctx, request)
		if err != nil || result == nil || !linksToBGG(result) {
			return result, err
		}
		result.Content = append(result.Content, assistantNote(answerGuide))
		return result, nil
	}
}

// assistantNote is tool result text meant for the model rather than the user.
// The audience annotation says so to clients that read it, and the note's own
// wording says so to the model in clients that don't.
func assistantNote(text string) mcp.TextContent {
	note := mcp.NewTextContent(text)
	note.Annotations = &mcp.Annotations{Audience: []mcp.Role{mcp.RoleAssistant}}
	return note
}

func linksToBGG(result *mcp.CallToolResult) bool {
	for _, content := range result.Content {
		if text, ok := content.(mcp.TextContent); ok && strings.Contains(text.Text, BGGHomeURL+"/") {
			return true
		}
	}
	return false
}
