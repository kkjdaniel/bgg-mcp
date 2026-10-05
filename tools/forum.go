package tools

import (
	"context"
	"html"
	"regexp"
	"strings"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/thread"
)

type ThreadDetailsResult struct {
	ID       int          `json:"id"`
	Subject  string       `json:"subject"`
	Link     string       `json:"link"`
	NumPosts int          `json:"num_posts"`
	Posts    []ThreadPost `json:"posts"`
}

type ThreadPost struct {
	ID     int    `json:"id"`
	Author string `json:"author"`
	Date   string `json:"date,omitempty"`
	Body   string `json:"body"`
}

var (
	lineBreakTagPattern = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>|</li>|</blockquote>`)
	htmlTagPattern      = regexp.MustCompile(`<[^>]*>`)
	inlineSpacePattern  = regexp.MustCompile(`[ \t\x{00a0}]+`)
	blankLinesPattern   = regexp.MustCompile(`\n{3,}`)
)

// cleanPostBody converts the HTML body of a forum post into plain text.
func cleanPostBody(body string) string {
	text := lineBreakTagPattern.ReplaceAllString(body, "\n")
	text = htmlTagPattern.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	text = inlineSpacePattern.ReplaceAllString(text, " ")

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	text = blankLinesPattern.ReplaceAllString(strings.Join(lines, "\n"), "\n\n")

	return strings.TrimSpace(text)
}

// extractThreadPosts converts thread articles to plain-text posts. A maxBodyLength
// of 0 leaves post bodies untruncated.
func extractThreadPosts(articles []thread.Article, maxBodyLength int) []ThreadPost {
	posts := make([]ThreadPost, len(articles))
	for i, article := range articles {
		body := cleanPostBody(article.Body)
		if runes := []rune(body); maxBodyLength > 0 && len(runes) > maxBodyLength {
			body = strings.TrimSpace(string(runes[:maxBodyLength])) + "…"
		}
		posts[i] = ThreadPost{
			ID:     article.ID,
			Author: article.Username,
			Date:   article.PostDate,
			Body:   body,
		}
	}
	return posts
}

// FetchThread returns a forum thread with every post converted to plain text.
func FetchThread(ctx context.Context, client *gogeek.Client, threadID int) (*ThreadDetailsResult, error) {
	threadDetail, err := thread.Query(ctx, client, threadID)
	if err != nil {
		return nil, err
	}

	return &ThreadDetailsResult{
		ID:       threadDetail.ID,
		Subject:  threadDetail.Subject,
		Link:     threadDetail.Link,
		NumPosts: threadDetail.NumArticles,
		Posts:    extractThreadPosts(threadDetail.Articles, 0),
	}, nil
}
