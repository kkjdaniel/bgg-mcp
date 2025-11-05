<p align="center">
  <img src="images/bgg-mcp-logo.png" width="200" alt="BGG MCP Logo">
</p>
<h1 align="center">BGG MCP: BoardGameGeek MCP Server</h1>

<p align="center">
  <a href="https://smithery.ai/server/@kkjdaniel/bgg-mcp"><img src="https://smithery.ai/badge/@kkjdaniel/bgg-mcp" alt="smithery badge"></a>
  <a href="https://archestra.ai/mcp-catalog/kkjdaniel__bgg-mcp"><img src="https://archestra.ai/mcp-catalog/api/badge/quality/kkjdaniel/bgg-mcp" alt="trust score badge"></a>
  <a href="https://github.com/modelcontextprotocol/registry"><img src="https://img.shields.io/badge/MCP_Registry-BGG_MCP-green" alt="MCP Registry"></a>
  <br>
  <a href="https://go.dev/"><img src="https://img.shields.io/github/go-mod/go-version/kkjdaniel/bgg-mcp" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/kkjdaniel/bgg-mcp" alt="License"></a>
  <a href="https://modelcontextprotocol.io"><img src="https://img.shields.io/badge/MCP-Protocol-blue" alt="MCP Protocol"></a>
</p>

BGG MCP provides access to the BoardGameGeek API through the [Model Context Protocol](https://www.anthropic.com/news/model-context-protocol), enabling retrieval and filtering of board game data, user collections, and profiles. The server is implemented in Go, using the [GoGeek](https://github.com/kkjdaniel/gogeek) library, which helps ensure robust API interactions.

Price data is provided by [BoardGamePrices.co.uk](https://boardgameprices.co.uk), offering real-time pricing from multiple retailers.

Game recommendations are powered by [Recommend.Games](https://recommend.games/), which provides algorithmic similarity recommendations based on BoardGameGeek data.

<a href="https://boardgamegeek.com/">
  <img src="images/powered-bgg.webp" width="160" alt="Powered by BGG">
</a>

## Demo

<div align="center">
  
  [![Rules Tool Demo Video](https://img.youtube.com/vi/cNX4WwVbFko/maxresdefault.jpg)](https://youtu.be/cNX4WwVbFko)
  
  **[▶️ Watch the Rules Tool Demo Video](https://youtu.be/cNX4WwVbFko)**
  
</div>

## Tools

### Core Tools

| Tool                 | Description                                                                 |
| -------------------- | --------------------------------------------------------------------------- |
| `bgg-search`         | Search for board games with type filtering (base games, expansions, or all) |
| `bgg-details`        | Get detailed information about a specific board game                        |
| `bgg-collection`     | Query and filter a user's game collection with extensive filtering options  |
| `bgg-hot`            | Get the current BGG hotness list                                            |
| `bgg-user`           | Get user profile information                                                |
| `bgg-price`          | Get current prices from multiple retailers using BGG IDs                    |
| `bgg-trade-finder`   | Find trading opportunities between two BGG users                            |
| `bgg-recommender`    | Get game recommendations based on similarity to a specific game             |
| `bgg-thread-details` | Get the full content of a specific BGG forum thread including all posts     |

### 🧪 Experimental Tools

| Tool        | Description                                                                                |
| ----------- | ------------------------------------------------------------------------------------------ |
| `bgg-rules` | Answer rules questions by searching BGG forums for relevant discussions and clarifications |

## Resources

BGG MCP exposes resources that AI assistants can access directly for contextual information:

| Resource              | URI                   | Description                                                       |
| --------------------- | --------------------- | ----------------------------------------------------------------- |
| `BGG Hotness`         | `bgg://hotness`       | Current BGG hotness list, always available                        |
| `My BGG Collection`   | `bgg://my-collection` | Your personal BGG collection (requires `BGG_USERNAME` to be set)  |

Resources provide AI assistants with direct access to BGG data without needing explicit tool calls, making conversations more natural and efficient.

## Prompts

- **Trade Sales Post** - Generate a formatted sales post for your BGG 'for trade' collection with discounted market prices
- **Game Recommendations** - Get personalized game recommendations based on your BGG collection and preferences

## Example Prompts

Here are some example prompts you can use to interact with the BGG MCP tools:

### 🔍 Search

```
"Search for Wingspan on BGG"
"How many expansions does Grand Austria Hotel have?"
"Search for Wingspan expansions only"
```

### 📊 Game Details

```
"Get details for Azul"
"Show me information about game ID 224517"
"What's the BGG rating for Gloomhaven?"
```

### 📚 Collection

```
"Show me ZeeGarcia's game collection"
"Show games rated 9+ in kkjdaniel's collection"
"List unplayed games in rahdo's collection"
"Find games for 6 players in kkjdaniel's collection"
"Show me all the games rated 3 and below in my collection"
"What games in my collection does rahdo want?"
"What games does kkjdaniel have that I want?"
```

### 🔥 Hotness

```
"Show me the current BGG hotness list"
"What's trending on BGG?"
```

### 👤 User Profile

```
"Show me details about BGG user rahdo"
"When did user ZeeGarcia join BGG?"
"How many buddies do I have on bgg?"
```

### 💰 Prices

```
"Get the best price for Wingspan in GBP"
"Show me the best UK price for Ark Nova"
"Compare prices for: Wingspan & Ark Nova"
```

### 🎯 Recommendations

```
"Recommend games similar to Wingspan"
"What games are like Azul but with at least 1000 ratings?"
"Find 5 games similar to Troyes"
```

### 📖 Rules (Experimental)

```
"[Your rules question about any board game] - use bgg-rules"
"How does [game mechanic] work in [game name]? use bgg-rules"
"Can I [specific action] in [game name]? use bgg-rules"
"What happens when [situation] in [game name]? use bgg-rules"
```

Note: Include "use bgg-rules" in your question to ensure the AI searches BGG forums for answers.

## Installation

> **Authentication Required**: Most BGG MCP tools require authentication to access BoardGameGeek's API. See the [Configuration section](#configuration) below for setup instructions.

You have multiple options for installing BGG MCP:

### A) Installing via Smithery (Recommended)

Get started in under a minute with [Smithery](https://smithery.ai/server/@kkjdaniel/bgg-mcp):

1. **Sign up** at Smithery and select your client (e.g., Claude Desktop)
2. **Click Configure** to set up your BGG authentication (see [Configuration](#configuration))
3. **Connect** and start using BGG tools immediately

#### Or via Connectors for Claude users

Add BGG MCP as a custom connector:

1. Go to **Settings → Connectors → Add custom connector**
2. Enter this URL:
   ```
   https://server.smithery.ai/@kkjdaniel/bgg-mcp/mcp
   ```
3. Click **Connect** to authorise

That's it! The server uses the latest Streamable HTTP transport.

**Tip:** Connectors added on Claude Desktop will also appear and work on mobile!

### B) MCP Registry

Install via the MCP Registry:

```bash
mcp install io.github.kkjdaniel/bgg-mcp
```

### C) Manual Setup

#### 1. Install Go

You will need to have Go installed on your system to build binary. This can be easily [downloaded and setup here](https://go.dev/doc/install), or you can use the package manager that you prefer such as Brew.

#### 2. Build

The project includes a Makefile to simplify building and managing the binary.

```bash
# Build the application (output goes to build/bgg-mcp)
make build

# Clean build artifacts
make clean

# Both clean and build
make all
```

Or you can simply build it directly with Go...

```bash
go build -o build/bgg-mcp
```

#### 3. Add MCP Config

In the `settings.json` (VS Code / Cursor) or `claude_desktop_config.json` add the following to your list of servers, pointing it to the binary you created earlier, once you load up your AI tool you should see the tools provided by the server connected:

```json
"bgg": {
    "command": "path/to/build/bgg-mcp",
    "args": ["-mode", "stdio"]
}
```

More details for configuring Claude can be [found here](https://modelcontextprotocol.io/quickstart/user).

## Configuration

### Authentication

BGG MCP v2.0+ uses the GoGeek v2.0 library which requires authentication for reliable access to BoardGameGeek's API.

You can configure authentication using **either** `BGG_API_KEY` (recommended) or `BGG_COOKIE`:

#### For Smithery Users

When connecting via [Smithery](https://smithery.ai/server/@kkjdaniel/bgg-mcp), you can configure authentication directly in the connection UI:

1. Click **Configure** button on the BGG MCP server page
2. Fill in the configuration form (under optional):
   - **BGG API Key** (recommended) - Get one at [boardgamegeek.com/applications](https://boardgamegeek.com/applications)
   - **BGG Cookie** (alternative) - Only if you're not using an API key
   - **BGG Username** - Your username for personalized features
3. Click **Save Configuration**, then connect with this profile

Your settings are stored securely and applied automatically when you use BGG MCP tools.

#### For Manual Setup

##### Option 1: API Key (Recommended)

Get an API key from [BoardGameGeek's API application form](https://boardgamegeek.com/applications) and add it to your configuration:

```json
"bgg": {
    "env": {
        "BGG_API_KEY": "your_api_key_here"
    }
}
```

##### Option 2: Cookie Authentication

Alternatively, you can use cookie-based authentication:

```json
"bgg": {
    "env": {
        "BGG_COOKIE": "bggusername=user; bggpassword=pass; SessionID=xyz"
    }
}
```

**Note**: If both are provided, `BGG_API_KEY` takes precedence.

### Username Configuration

You can optionally set the `BGG_USERNAME` environment variable to enable "me" and "my" references in queries:

```json
"bgg": {
    "env": {
        "BGG_USERNAME": "your_bgg_username",
        "BGG_API_KEY": "your_api_key_here"
    }
}
```

This enables:

- **Collection queries**: "Show my collection" instead of specifying your username
- **User queries**: "Show my BGG profile"
- **AI assistance**: The AI can automatically use your username for comparisons and analysis

**Note**: When you use self-references (me, my, I) without setting BGG_USERNAME, you'll get a clear error message.
