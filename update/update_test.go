package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"v1.9.0", "1.8.0", true},
		{"v1.8.1", "1.8.0", true},
		{"v2.0.0", "1.12.3", true},
		{"v1.10.0", "1.9.0", true},
		{"v1.8.0", "1.8.0", false},
		{"v1.7.9", "1.8.0", false},
		{"v1.9.0-rc1", "1.8.0", false},
		{"v1.9", "1.8.0", false},
		{"", "1.8.0", false},
		{"v1.9.0", "dev", false},
	}
	for _, c := range cases {
		if got := newer(c.a, c.b); got != c.want {
			t.Errorf("newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// fakeGitHub serves tag as the latest release, or fails while tag is "".
func fakeGitHub(t *testing.T, tag *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *tag == "" {
			http.Error(w, "rate limited", http.StatusForbidden)
			return
		}
		w.Write([]byte(`{"tag_name":"` + *tag + `","html_url":"https://github.com/kkjdaniel/bgg-mcp/releases/tag/` + *tag + `"}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestCheck(t *testing.T) {
	tag := "v1.9.0"
	c := NewChecker("1.8.0")
	c.url = fakeGitHub(t, &tag).URL

	c.check(context.Background())
	notice := c.Notice()
	for _, want := range []string{"v1.9.0", "v1.8.0", "https://github.com/kkjdaniel/bgg-mcp/releases/tag/v1.9.0"} {
		if !strings.Contains(notice, want) {
			t.Errorf("notice %q should mention %q", notice, want)
		}
	}

	tag = ""
	c.check(context.Background())
	if c.Notice() != notice {
		t.Errorf("a failed check should keep the last notice, got %q", c.Notice())
	}

	tag = "v1.8.0"
	c.check(context.Background())
	if c.Notice() != "" {
		t.Errorf("no notice once the server is on the latest release, got %q", c.Notice())
	}
}

func TestMiddleware(t *testing.T) {
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(`{"name":"Catan"}`), nil
	}

	c := NewChecker("1.8.0")
	result, _ := c.Middleware(handler)(context.Background(), mcp.CallToolRequest{})
	if len(result.Content) != 1 {
		t.Errorf("no notice should be added while up to date, got %d content items", len(result.Content))
	}

	c.notice = "Update available"
	result, _ = c.Middleware(handler)(context.Background(), mcp.CallToolRequest{})
	if len(result.Content) != 2 {
		t.Fatalf("the notice should follow the result, got %d content items", len(result.Content))
	}
	if text := result.Content[0].(mcp.TextContent).Text; text != `{"name":"Catan"}` {
		t.Errorf("the result itself should be unchanged, got %q", text)
	}
	notice := result.Content[1].(mcp.TextContent)
	if notice.Text != "Update available" {
		t.Errorf("notice = %q", notice.Text)
	}
}
