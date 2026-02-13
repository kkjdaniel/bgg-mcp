package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kkjdanie/bgg-mcp/prompts"
	"github.com/kkjdanie/bgg-mcp/resources"
	"github.com/kkjdanie/bgg-mcp/tools"
	"github.com/kkjdaniel/gogeek/v2"
	"github.com/mark3labs/mcp-go/server"
)

func initializeGoGeekClient() *gogeek.Client {
	return createClientFromEnv()
}

func createClientFromEnv() *gogeek.Client {
	if apiKey := os.Getenv("BGG_API_KEY"); apiKey != "" {
		return gogeek.NewClient(gogeek.WithAPIKey(apiKey))
	}

	if cookie := os.Getenv("BGG_COOKIE"); cookie != "" {
		return gogeek.NewClient(gogeek.WithCookie(cookie))
	}

	return gogeek.NewClient()
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
		return gogeek.NewClient(gogeek.WithAPIKey(apiKey))
	}
	if cookie != "" {
		return gogeek.NewClient(gogeek.WithCookie(cookie))
	}
	return gogeek.NewClient()
}

func createMCPServer(client *gogeek.Client) *server.MCPServer {
	s := server.NewMCPServer(
		"BGG MCP",
		"1.6.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	detailsTool, detailsHandler := tools.DetailsTool(client)
	s.AddTool(detailsTool, detailsHandler)

	collectionTool, collectionHandler := tools.CollectionTool(client)
	s.AddTool(collectionTool, collectionHandler)

	hotnessTool, hotnessHandler := tools.HotnessTool(client)
	s.AddTool(hotnessTool, hotnessHandler)

	userTool, userHandler := tools.UserTool(client)
	s.AddTool(userTool, userHandler)

	searchTool, searchHandler := tools.SearchTool(client)
	s.AddTool(searchTool, searchHandler)

	priceTool, priceHandler := tools.PriceTool()
	s.AddTool(priceTool, priceHandler)

	tradeFinderTool, tradeFinderHandler := tools.TradeFinderTool(client)
	s.AddTool(tradeFinderTool, tradeFinderHandler)

	recommenderTool, recommenderHandler := tools.RecommenderTool(client)
	s.AddTool(recommenderTool, recommenderHandler)

	rulesTool, rulesHandler := tools.RulesTool(client)
	s.AddTool(rulesTool, rulesHandler)

	threadDetailsTool, threadDetailsHandler := tools.ThreadDetailsTool(client)
	s.AddTool(threadDetailsTool, threadDetailsHandler)

	hotnessResource, hotnessResourceHandler := resources.HotnessResource(client)
	s.AddResource(hotnessResource, hotnessResourceHandler)

	myCollectionResource, myCollectionResourceHandler := resources.MyCollectionResource(client)
	s.AddResource(myCollectionResource, myCollectionResourceHandler)

	prompts.RegisterPrompts(s)

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
		mcpServer := createMCPServer(client)
		runStdioServer(mcpServer)
	default:
		log.Fatalf("Invalid mode: %s. Use 'stdio' or 'http'", mode)
	}
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

		originalUsername := os.Getenv("BGG_USERNAME")
		if username != "" {
			os.Setenv("BGG_USERNAME", username)
			defer os.Setenv("BGG_USERNAME", originalUsername)
		}

		sessionMCPServer := createMCPServer(sessionClient)

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
