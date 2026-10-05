package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kkjdaniel/bgg-mcp/tools"
	"github.com/kkjdaniel/gogeek/v3"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func jsonContents(uri string, text string) []mcp.ResourceContents {
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     text,
		},
	}
}

// templateArgument reads a variable matched from a resource template URI.
func templateArgument(request mcp.ReadResourceRequest, name string) (string, error) {
	var value string
	switch v := request.Params.Arguments[name].(type) {
	case string:
		value = v
	case []string:
		if len(v) > 0 {
			value = v[0]
		}
	}

	if value == "" {
		return "", fmt.Errorf("missing %s in resource URI %s", name, request.Params.URI)
	}
	return value, nil
}

func templateIDArgument(request mcp.ReadResourceRequest, name string) (int, error) {
	value, err := templateArgument(request, name)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(value)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid %s in resource URI %s", name, request.Params.URI)
	}
	return id, nil
}

func UserCollectionTemplate(client *gogeek.Client) (mcp.ResourceTemplate, server.ResourceTemplateHandlerFunc) {
	template := mcp.NewResourceTemplate(
		"bgg://collection/{username}",
		"BGG User Collection",
		mcp.WithTemplateDescription("The owned games in a BoardGameGeek user's collection, with their ratings, play counts, and status"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		username, err := templateArgument(request, "username")
		if err != nil {
			return nil, err
		}

		return ownedCollectionContents(ctx, client, request.Params.URI, username)
	}

	return template, handler
}

func GameTemplate(client *gogeek.Client) (mcp.ResourceTemplate, server.ResourceTemplateHandlerFunc) {
	template := mcp.NewResourceTemplate(
		"bgg://game/{id}",
		"BGG Game",
		mcp.WithTemplateDescription("Essential information about a board game by its BoardGameGeek ID: description, mechanics, player count, playtime, complexity, and ratings"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		gameID, err := templateIDArgument(request, "id")
		if err != nil {
			return nil, err
		}

		info, err := tools.FetchGameInfo(ctx, client, gameID)
		if err != nil {
			return nil, fmt.Errorf("error fetching game: %v", err)
		}

		out, err := json.Marshal(info)
		if err != nil {
			return nil, fmt.Errorf("error formatting results: %v", err)
		}

		return jsonContents(request.Params.URI, string(out)), nil
	}

	return template, handler
}

func ThreadTemplate(client *gogeek.Client) (mcp.ResourceTemplate, server.ResourceTemplateHandlerFunc) {
	template := mcp.NewResourceTemplate(
		"bgg://thread/{id}",
		"BGG Forum Thread",
		mcp.WithTemplateDescription("A BoardGameGeek forum thread by its ID, with every post as plain text"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	handler := func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		threadID, err := templateIDArgument(request, "id")
		if err != nil {
			return nil, err
		}

		result, err := tools.FetchThread(ctx, client, threadID)
		if err != nil {
			return nil, fmt.Errorf("error fetching thread: %v", err)
		}

		out, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("error formatting results: %v", err)
		}

		return jsonContents(request.Params.URI, string(out)), nil
	}

	return template, handler
}
