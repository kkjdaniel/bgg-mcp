// Package update tells users when a newer BGG MCP release is out. It asks
// GitHub for the latest release in the background and adds a short notice to
// tool results while the running server is behind it.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	latestReleaseURL = "https://api.github.com/repos/kkjdaniel/bgg-mcp/releases/latest"
	releasesPage     = "https://github.com/kkjdaniel/bgg-mcp/releases/latest"

	// checkInterval is how long one check holds. A desktop client can keep the
	// server running for days, so it looks again rather than only at startup.
	checkInterval = 24 * time.Hour
)

// Checker remembers whether GitHub's latest release is newer than the running
// version. It starts out with no notice, and keeps the last one it had when a
// check fails, so being offline never adds or clears a notice.
type Checker struct {
	current string
	url     string
	client  *http.Client

	mu     sync.RWMutex
	notice string
}

// NewChecker returns a Checker for the running version, such as "1.8.0".
func NewChecker(current string) *Checker {
	return &Checker{
		current: current,
		url:     latestReleaseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Start checks for a newer release now and then once a day until ctx ends. It
// returns straight away, so a slow or unreachable GitHub never holds up a tool.
func (c *Checker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		for {
			c.check(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// Notice is the message for the model while a newer release is out, or "".
func (c *Checker) Notice() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.notice
}

// ToolNote is added to every tool's description while the update check runs.
// The notice in the results is a plain fact, because models rightly treat
// instructions inside a tool result as untrusted, so what to do with it is
// said here, in the tool definition.
const ToolNote = "If the result ends with a line saying a newer version of BGG MCP is available, tell the user once in the conversation, with its link."

// Middleware adds the notice to every tool result while a newer release is out.
// A desktop client keeps one server running across many conversations, so every
// result carries it, and ToolNote asks the model to mention it once in each.
func (c *Checker) Middleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := next(ctx, request)
		if err != nil || result == nil {
			return result, err
		}
		if notice := c.Notice(); notice != "" {
			result.Content = append(result.Content, mcp.NewTextContent(notice))
		}
		return result, nil
	}
}

type release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func (c *Checker) check(ctx context.Context) {
	latest, err := c.latestRelease(ctx)
	if err != nil {
		log.Printf("Update check failed: %v", err)
		return
	}

	notice := ""
	if newer(latest.TagName, c.current) {
		link := latest.HTMLURL
		if link == "" {
			link = releasesPage
		}
		notice = fmt.Sprintf("A newer version of BGG MCP is available: %s (this server is running v%s). Release notes and downloads: %s",
			latest.TagName, strings.TrimPrefix(c.current, "v"), link)
	}

	c.mu.Lock()
	c.notice = notice
	c.mu.Unlock()
}

func (c *Checker) latestRelease(ctx context.Context) (release, error) {
	var latest release

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return latest, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "bgg-mcp/"+c.current)

	resp, err := c.client.Do(req)
	if err != nil {
		return latest, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return latest, fmt.Errorf("GitHub returned %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return latest, fmt.Errorf("reading the latest release: %w", err)
	}
	return latest, nil
}

// newer reports whether version a is later than version b. Both must be
// "major.minor.patch", with or without a leading "v". Anything else, such as a
// pre-release suffix, is never newer, so an odd tag can't raise a false notice.
func newer(a, b string) bool {
	va, okA := parseVersion(a)
	vb, okB := parseVersion(b)
	if !okA || !okB {
		return false
	}
	for i := range va {
		if va[i] != vb[i] {
			return va[i] > vb[i]
		}
	}
	return false
}

func parseVersion(v string) ([3]int, bool) {
	var parsed [3]int
	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(parts) != len(parsed) {
		return parsed, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return parsed, false
		}
		parsed[i] = n
	}
	return parsed, true
}
