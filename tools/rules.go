package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/forum"
	"github.com/kkjdaniel/gogeek/v3/forumlist"
	"github.com/kkjdaniel/gogeek/v3/thread"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	rulesDefaultLimit      = 10
	rulesMaxLimit          = 50
	rulesMaxForumPages     = 10
	rulesForumPageSize     = 50
	rulesMaxReferences     = 5
	rulesInlineThreads     = 2
	rulesInlinePosts       = 10
	rulesInlinePostMaxSize = 1500
)

type RulesThread struct {
	ID             int          `json:"id"`
	Subject        string       `json:"subject"`
	Replies        int          `json:"replies"`
	LastPost       string       `json:"last_post,omitempty"`
	Link           string       `json:"link"`
	Posts          []ThreadPost `json:"posts,omitempty"`
	PostsTruncated bool         `json:"posts_truncated,omitempty"`
}

type RulesResult struct {
	GameName         string        `json:"game_name,omitempty"`
	GameID           int           `json:"game_id"`
	ForumTitle       string        `json:"forum_title"`
	TotalThreads     int           `json:"total_threads"`
	ThreadsSearched  int           `json:"threads_searched"`
	Question         string        `json:"question,omitempty"`
	Matches          []RulesThread `json:"matches"`
	ReferenceThreads []RulesThread `json:"reference_threads,omitempty"`
	Guidance         string        `json:"guidance"`
}

var (
	wordPattern            = regexp.MustCompile(`[a-z0-9]+`)
	referenceThreadPattern = regexp.MustCompile(`(?i)\b(faqs?|f\.a\.q|errata|official|frequently asked|common(ly asked)? questions)\b`)
)

var rulesStopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true,
	"can": true, "could": true, "do": true, "doe": true, "does": true, "for": true, "from": true,
	"game": true, "happen": true, "have": true, "how": true, "i": true, "if": true, "in": true,
	"is": true, "it": true, "my": true, "of": true, "on": true, "or": true, "question": true,
	"rule": true, "should": true, "that": true, "the": true, "their": true, "there": true,
	"they": true, "thi": true, "this": true, "to": true, "wa": true, "was": true, "we": true,
	"what": true, "when": true, "where": true, "which": true, "who": true, "why": true,
	"will": true, "with": true, "work": true, "would": true, "you": true, "your": true,
}

func stemWord(word string) string {
	for _, suffix := range []string{"ing", "ed", "es", "s"} {
		if strings.HasSuffix(word, suffix) && len(word)-len(suffix) >= 3 {
			return strings.TrimSuffix(word, suffix)
		}
	}
	return word
}

// rulesTerms splits text into the distinct, stemmed terms that are useful for matching.
func rulesTerms(text string) map[string]bool {
	terms := map[string]bool{}
	for _, word := range wordPattern.FindAllString(strings.ToLower(text), -1) {
		stem := stemWord(word)
		if rulesStopWords[word] || rulesStopWords[stem] {
			continue
		}
		terms[stem] = true
	}
	return terms
}

// rankThreads returns the threads whose subjects match the question, most relevant
// first. Terms that are rare across the forum count for more than common ones, and
// terms from the game's own name are ignored.
func rankThreads(threads []forum.Thread, question string, gameName string) []forum.Thread {
	queryTerms := rulesTerms(question)
	for term := range rulesTerms(gameName) {
		delete(queryTerms, term)
	}
	if len(queryTerms) == 0 {
		return nil
	}

	subjectTerms := make([]map[string]bool, len(threads))
	documentFrequency := map[string]int{}
	for i, t := range threads {
		subjectTerms[i] = rulesTerms(t.Subject)
		for term := range subjectTerms[i] {
			documentFrequency[term]++
		}
	}

	type scoredThread struct {
		thread forum.Thread
		score  float64
	}
	scored := []scoredThread{}
	for i, t := range threads {
		score := 0.0
		for term := range queryTerms {
			if subjectTerms[i][term] {
				score += math.Log(1 + float64(len(threads))/float64(documentFrequency[term]))
			}
		}
		if score == 0 {
			continue
		}
		score += 0.1 * math.Log(1+float64(t.NumArticles))
		scored = append(scored, scoredThread{thread: t, score: score})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	ranked := make([]forum.Thread, len(scored))
	for i, s := range scored {
		ranked[i] = s.thread
	}
	return ranked
}

// findReferenceThreads returns FAQ, errata and other official-sounding threads,
// which often answer questions that their titles do not mention.
func findReferenceThreads(threads []forum.Thread, exclude map[int]bool) []forum.Thread {
	references := []forum.Thread{}
	for _, t := range threads {
		if !exclude[t.ID] && referenceThreadPattern.MatchString(t.Subject) {
			references = append(references, t)
		}
	}

	sort.SliceStable(references, func(i, j int) bool {
		return references[i].NumArticles > references[j].NumArticles
	})

	if len(references) > rulesMaxReferences {
		references = references[:rulesMaxReferences]
	}
	return references
}

func toRulesThread(t forum.Thread) RulesThread {
	return RulesThread{
		ID:       t.ID,
		Subject:  t.Subject,
		Replies:  max(t.NumArticles-1, 0),
		LastPost: t.LastPostDate,
		Link:     fmt.Sprintf("https://boardgamegeek.com/thread/%d", t.ID),
	}
}

func toRulesThreads(threads []forum.Thread) []RulesThread {
	result := make([]RulesThread, len(threads))
	for i, t := range threads {
		result[i] = toRulesThread(t)
	}
	return result
}

func findRulesForum(forums []forumlist.Forum) *forumlist.Forum {
	var fallback *forumlist.Forum
	for i := range forums {
		title := strings.ToLower(strings.TrimSpace(forums[i].Title))
		if title == "rules" {
			return &forums[i]
		}
		if fallback == nil && strings.Contains(title, "rules") {
			fallback = &forums[i]
		}
	}
	return fallback
}

func RulesTool(client *gogeek.Client) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-rules",
		mcp.WithDescription("Use this tool when users ask rules questions about board games (e.g., 'How does X work?', 'Can I do Y?', 'What happens when Z?'). Searches the game's BoardGameGeek rules forum for the threads most relevant to the question and returns them ranked, with the posts of the top matches included."),
		mcp.WithString("name",
			mcp.Description("The name of the board game. Use when the BGG ID is not known."),
		),
		mcp.WithNumber("id",
			mcp.Description("The BoardGameGeek ID of the board game. Preferred over 'name' when already known."),
		),
		mcp.WithString("question",
			mcp.Description("The rules question, or its key terms (e.g., 'can I trade resources during another player's turn'). Used to rank forum threads by relevance. Strongly recommended: without it only the most recent threads are returned."),
		),
		mcp.WithNumber("limit",
			mcp.Description(fmt.Sprintf("Maximum number of matching threads to return (default: %d, max: %d)", rulesDefaultLimit, rulesMaxLimit)),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()

		var gameID int
		var gameName string
		var err error

		if idVal, ok := arguments["id"]; ok && idVal != nil {
			switch v := idVal.(type) {
			case float64:
				gameID = int(v)
			case string:
				gameID, err = strconv.Atoi(v)
				if err != nil {
					return mcp.NewToolResultText("Invalid game ID format"), nil
				}
			}
		} else if nameVal, ok := arguments["name"].(string); ok && nameVal != "" {
			bestMatch, err := findBestGameMatch(ctx, client, nameVal)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Failed to find game: %v", err)), nil
			}
			gameID = bestMatch.ID
			gameName = bestMatch.Name.Value
		} else {
			return mcp.NewToolResultText("Either 'name' or 'id' parameter is required"), nil
		}

		question, _ := arguments["question"].(string)
		question = strings.TrimSpace(question)

		limit := rulesDefaultLimit
		if limitVal, ok := arguments["limit"].(float64); ok && limitVal >= 1 {
			limit = min(int(limitVal), rulesMaxLimit)
		}

		forums, err := forumlist.Query(ctx, client, gameID, forumlist.Thing)
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Failed to get forum list: %v", err)), nil
		}

		rulesForum := findRulesForum(forums.Forums)
		if rulesForum == nil {
			return mcp.NewToolResultText(fmt.Sprintf("No rules forum found for game ID %d", gameID)), nil
		}

		result := RulesResult{
			GameName:   gameName,
			GameID:     gameID,
			ForumTitle: rulesForum.Title,
			Question:   question,
		}

		// Without a question there is nothing to rank, so only the most recent page is needed
		maxPages := rulesMaxForumPages
		if question == "" {
			maxPages = 1
		}

		allThreads := []forum.Thread{}
		for page := 1; page <= maxPages; page++ {
			forumPage, err := forum.Query(ctx, client, rulesForum.ID, forum.WithPage(page))
			if err != nil {
				// Earlier pages are still worth searching if a later one fails
				if page > 1 {
					break
				}
				return mcp.NewToolResultText(fmt.Sprintf("Failed to get rules forum threads: %v", err)), nil
			}

			if page == 1 {
				result.TotalThreads = forumPage.NumThreads
			}

			allThreads = append(allThreads, forumPage.Threads...)

			if len(forumPage.Threads) < rulesForumPageSize {
				break
			}
		}
		result.ThreadsSearched = len(allThreads)

		var matches []forum.Thread
		if question != "" {
			matches = rankThreads(allThreads, question, gameName)
		}
		ranked := len(matches) > 0

		switch {
		case ranked:
			result.Guidance = "Matches are ranked by how well their titles fit the question, and the top matches include their posts. Answer from those posts where they settle the question, and cite the thread link. Use bgg-thread-details to read any other thread, including reference threads (FAQ/errata), which often cover questions their titles do not mention. Check that the game is the one the user meant."
		case question != "":
			matches = allThreads
			result.Guidance = "No thread titles matched the question, so the most recent threads are listed instead. Try again with different key terms, or use bgg-thread-details on a reference thread (FAQ/errata) or any thread that looks relevant."
		default:
			matches = allThreads
			result.Guidance = "No question was given, so the most recent threads are listed. Call again with the 'question' parameter to search the whole forum by relevance, or use bgg-thread-details to read a thread."
		}

		if len(matches) > limit {
			matches = matches[:limit]
		}
		result.Matches = toRulesThreads(matches)

		matchedIDs := map[int]bool{}
		for _, t := range matches {
			matchedIDs[t.ID] = true
		}
		result.ReferenceThreads = toRulesThreads(findReferenceThreads(allThreads, matchedIDs))

		if ranked {
			for i := 0; i < len(result.Matches) && i < rulesInlineThreads; i++ {
				threadDetail, err := thread.Query(ctx, client, result.Matches[i].ID, thread.WithCount(rulesInlinePosts))
				if err != nil {
					continue
				}
				result.Matches[i].Posts = extractThreadPosts(threadDetail.Articles, rulesInlinePostMaxSize)
				result.Matches[i].PostsTruncated = threadDetail.NumArticles > len(threadDetail.Articles)
			}
		}

		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Failed to format result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}

	return tool, handler
}
