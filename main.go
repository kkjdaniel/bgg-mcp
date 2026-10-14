package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kkjdaniel/bgg-mcp/prompts"
	"github.com/kkjdaniel/bgg-mcp/resources"
	"github.com/kkjdaniel/bgg-mcp/tools"
	"github.com/kkjdaniel/bgg-mcp/update"
	"github.com/kkjdaniel/gogeek/v3"
	"github.com/mark3labs/mcp-go/server"
)

func initializeGoGeekClient() *gogeek.Client {
	return createClientFromEnv()
}

func createClientFromEnv() *gogeek.Client {
	if apiKey := os.Getenv("BGG_API_KEY"); apiKey != "" {
		return gogeek.NewClient(gogeek.APIKey(apiKey))
	}

	if cookie := os.Getenv("BGG_COOKIE"); cookie != "" {
		return gogeek.NewClient(gogeek.Cookie(cookie))
	}

	return gogeek.NewClient(gogeek.Auth{})
}

func parseSessionConfig(r *http.Request) (apiKey, cookie, username string) {
	query := r.URL.Query()
	apiKey = query.Get("BGG_API_KEY")
	cookie = query.Get("BGG_COOKIE")
	username = query.Get("BGG_USERNAME")
	return
}

func createClientFromSessionConfig(apiKey, cookie string) *gogeek.Client {
	if apiKey != "" {
		return gogeek.NewClient(gogeek.APIKey(apiKey))
	}
	if cookie != "" {
		return gogeek.NewClient(gogeek.Cookie(cookie))
	}
	return gogeek.NewClient(gogeek.Auth{})
}

// serverVersion is the version the server reports, and the one the update check
// compares with the latest GitHub release.
const serverVersion = "1.9.0"

// serverInstructions asks the model to show users the boardgamegeek.com links
// that come with BGG data, which BoardGameGeek asks of answers that use it, and
// to lay answers out so that long lists of games stay easy to read.
const serverInstructions = `BoardGameGeek (BGG) asks that answers using its data link back to boardgamegeek.com, so always show the user the links that come with this server's results:

- Games, designers, publishers, categories, mechanics, users, guilds and forum threads in the results carry a "link" (or a field ending in "_link") to their BGG page. When your answer names one of them, put its link on the name as a Markdown link, such as [Brass: Birmingham](https://boardgamegeek.com/boardgame/224517), rather than showing a bare URL.
- When your answer uses BGG data without naming anything that has a link, such as how many games a collection holds, link the page the data came from (such as the "collection_link" of a collection), or else https://boardgamegeek.com.
- Use the links exactly as the results give them rather than writing BGG URLs from memory.

Lay answers out to be scanned rather than read as prose:

- Start with a one-line answer, such as how many games matched, then give the detail.
- Never run more than a few games together in a sentence or a comma-separated paragraph.
- For a broad list, such as every game in a collection that matches, use a Markdown table with one row per game and a column for each figure the question turns on, such as the personal rating, best player counts or play time. Bullets suit a handful of games at most.
- When a list runs long, split it into tables under short headings by whatever answers the question, such as rating, best player count, or base games and expansions.
- For a single game, put its key figures (such as players, best player count, play time, complexity, rating and rank) in a compact table rather than a paragraph. Lay out marketplace listings and editions as tables too.
- Answer from the figures in the results, such as "best_players" for which games play best at a player count, rather than from memory. Keep caveats to a short line at the end, and only include ones that change how the answer should be read.`

// createMCPServer builds the server. updates adds a notice to tool results while
// a newer release is out, and is nil when nothing should be added.
func createMCPServer(client *gogeek.Client, username string, updates *update.Checker) *server.MCPServer {
	options := []server.ServerOption{
		server.WithInstructions(serverInstructions),
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
		server.WithRecovery(),
	}
	if updates != nil {
		options = append(options, server.WithToolHandlerMiddleware(updates.Middleware))
	}
	s := server.NewMCPServer("BGG MCP", serverVersion, options...)

	detailsTool, detailsHandler := tools.DetailsTool(client)
	s.AddTool(detailsTool, detailsHandler)

	collectionTool, collectionHandler := tools.CollectionTool(client, username)
	s.AddTool(collectionTool, collectionHandler)

	hotnessTool, hotnessHandler := tools.HotnessTool(client)
	s.AddTool(hotnessTool, hotnessHandler)

	userTool, userHandler := tools.UserTool(client, username)
	s.AddTool(userTool, userHandler)

	searchTool, searchHandler := tools.SearchTool(client)
	s.AddTool(searchTool, searchHandler)

	priceTool, priceHandler := tools.PriceTool()
	s.AddTool(priceTool, priceHandler)

	tradeFinderTool, tradeFinderHandler := tools.TradeFinderTool(client, username)
	s.AddTool(tradeFinderTool, tradeFinderHandler)

	recommenderTool, recommenderHandler := tools.RecommenderTool(client)
	s.AddTool(recommenderTool, recommenderHandler)

	rulesTool, rulesHandler := tools.RulesTool(client)
	s.AddTool(rulesTool, rulesHandler)

	threadDetailsTool, threadDetailsHandler := tools.ThreadDetailsTool(client)
	s.AddTool(threadDetailsTool, threadDetailsHandler)

	hotnessResource, hotnessResourceHandler := resources.HotnessResource(client)
	s.AddResource(hotnessResource, hotnessResourceHandler)

	if username != "" {
		myCollectionResource, myCollectionResourceHandler := resources.MyCollectionResource(client, username)
		s.AddResource(myCollectionResource, myCollectionResourceHandler)
	}

	userCollectionTemplate, userCollectionTemplateHandler := resources.UserCollectionTemplate(client)
	s.AddResourceTemplate(userCollectionTemplate, userCollectionTemplateHandler)

	gameTemplate, gameTemplateHandler := resources.GameTemplate(client)
	s.AddResourceTemplate(gameTemplate, gameTemplateHandler)

	threadTemplate, threadTemplateHandler := resources.ThreadTemplate(client)
	s.AddResourceTemplate(threadTemplate, threadTemplateHandler)

	prompts.RegisterPrompts(s, username)

	return s
}

func main() {
	var mode string
	var port string
	
	flag.StringVar(&mode, "mode", "stdio", "Server mode: stdio or http")
	flag.StringVar(&port, "port", "8080", "Port for HTTP server (only used in http mode)")
	flag.Parse()

	if envMode := os.Getenv("MCP_MODE"); envMode != "" {
		mode = envMode
	}
	
	if envPort := os.Getenv("MCP_PORT"); envPort != "" {
		port = envPort
	}

	switch mode {
	case "http":
		runHTTPServer(port)
	case "stdio":
		client := initializeGoGeekClient()
		mcpServer := createMCPServer(client, os.Getenv("BGG_USERNAME"), startUpdateCheck())
		runStdioServer(mcpServer)
	default:
		log.Fatalf("Invalid mode: %s. Use 'stdio' or 'http'", mode)
	}
}

// startUpdateCheck starts looking for a newer release, unless the user turned
// that off with BGG_DISABLE_UPDATE_CHECK. Only stdio mode checks: the users of
// an HTTP deployment can't update it themselves.
func startUpdateCheck() *update.Checker {
	if disabled, _ := strconv.ParseBool(os.Getenv("BGG_DISABLE_UPDATE_CHECK")); disabled {
		return nil
	}
	updates := update.NewChecker(serverVersion)
	updates.Start(context.Background())
	return updates
}

func runStdioServer(mcpServer *server.MCPServer) {
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("STDIO server error: %v", err)
	}
}

func runHTTPServer(port string) {
	baseURL := os.Getenv("MCP_BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%s", port)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/.well-known/mcp-config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		configSchema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "BGG_API_KEY": {
      "type": "string",
      "title": "BGG API Key (Recommended)",
      "description": "API key from BoardGameGeek for authentication. Get one at https://boardgamegeek.com/applications"
    },
    "BGG_COOKIE": {
      "type": "string",
      "title": "BGG Cookie (Alternative)",
      "description": "Cookie string for BGG authentication. Only needed if not using API key"
    },
    "BGG_USERNAME": {
      "type": "string",
      "title": "BGG Username",
      "description": "Your BGG username for personalized features"
    }
  }
}`
		w.Write([]byte(configSchema))
	})

	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		apiKey, cookie, username := parseSessionConfig(r)

		var sessionClient *gogeek.Client
		if apiKey != "" || cookie != "" {
			sessionClient = createClientFromSessionConfig(apiKey, cookie)
			log.Printf("Using session configuration for request")
		} else {
			sessionClient = createClientFromEnv()
		}

		if username == "" {
			username = os.Getenv("BGG_USERNAME")
		}

		sessionMCPServer := createMCPServer(sessionClient, username, nil)

		httpServer := server.NewStreamableHTTPServer(sessionMCPServer,
			server.WithEndpointPath("/mcp"),
			server.WithStateLess(true),
			server.WithHeartbeatInterval(30*time.Second),
		)

		httpServer.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down HTTP server...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("HTTP endpoint: %s/mcp", baseURL)
	log.Printf("Supports session configuration via query parameters")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
