package tools

import (
	"testing"

	"github.com/kkjdaniel/gogeek/v3/forum"
	"github.com/kkjdaniel/gogeek/v3/forumlist"
)

func TestRankThreads(t *testing.T) {
	threads := []forum.Thread{
		{ID: 1, Subject: "Wingspan card storage ideas", NumArticles: 40},
		{ID: 2, Subject: "Do pink powers activate between turns?", NumArticles: 6},
		{ID: 3, Subject: "Tucking cards and end of round goals", NumArticles: 3},
		{ID: 4, Subject: "Pink power timing with multiple players", NumArticles: 12},
		{ID: 5, Subject: "Official FAQ", NumArticles: 90},
	}

	ranked := rankThreads(threads, "When do pink powers trigger in Wingspan?", "Wingspan")
	if len(ranked) != 2 {
		t.Fatalf("expected 2 matching threads, got %d: %+v", len(ranked), ranked)
	}
	for _, got := range ranked {
		if got.ID != 2 && got.ID != 4 {
			t.Errorf("unexpected thread matched: %+v", got)
		}
	}

	if got := rankThreads(threads, "how does this work?", "Wingspan"); len(got) != 0 {
		t.Errorf("expected no matches for a question with only stop words, got %+v", got)
	}
}

func TestRankThreadsPrefersRareTerms(t *testing.T) {
	threads := []forum.Thread{
		{ID: 1, Subject: "Card draw question", NumArticles: 2},
		{ID: 2, Subject: "Card limit", NumArticles: 2},
		{ID: 3, Subject: "Card trading", NumArticles: 2},
		{ID: 4, Subject: "Robber and the longest road", NumArticles: 2},
	}

	ranked := rankThreads(threads, "does the robber block a card", "")
	if len(ranked) == 0 || ranked[0].ID != 4 {
		t.Fatalf("expected the robber thread first, got %+v", ranked)
	}
}

func TestFindReferenceThreads(t *testing.T) {
	threads := []forum.Thread{
		{ID: 1, Subject: "Official FAQ", NumArticles: 90},
		{ID: 2, Subject: "Errata for first printing", NumArticles: 10},
		{ID: 3, Subject: "Unofficially my favourite game", NumArticles: 200},
		{ID: 4, Subject: "Rules clarification on scoring", NumArticles: 150},
		{ID: 5, Subject: "Is this ruling correct?", NumArticles: 120},
		{ID: 6, Subject: "Frequently asked questions (updated)", NumArticles: 30},
		{ID: 7, Subject: "Common questions answered", NumArticles: 20},
		{ID: 8, Subject: "Unofficial FAQs", NumArticles: 5},
		{ID: 9, Subject: "Designer FAQ", NumArticles: 300},
	}

	references := findReferenceThreads(threads, map[int]bool{9: true})
	wantIDs := []int{1, 6, 7, 2, 8}
	if len(references) != len(wantIDs) {
		t.Fatalf("expected %d reference threads, got %+v", len(wantIDs), references)
	}
	for i, want := range wantIDs {
		if references[i].ID != want {
			t.Errorf("reference %d: expected thread %d, got %+v", i, want, references[i])
		}
	}
}

func TestFindRulesForum(t *testing.T) {
	forums := []forumlist.Forum{
		{ID: 1, Title: "General"},
		{ID: 2, Title: "House Rules"},
		{ID: 3, Title: "Rules"},
	}
	if got := findRulesForum(forums); got == nil || got.ID != 3 {
		t.Fatalf("expected the exact Rules forum, got %+v", got)
	}
	if got := findRulesForum(forums[:2]); got == nil || got.ID != 2 {
		t.Fatalf("expected a fallback forum containing 'rules', got %+v", got)
	}
	if got := findRulesForum(forums[:1]); got != nil {
		t.Fatalf("expected no rules forum, got %+v", got)
	}
}

func TestCleanPostBody(t *testing.T) {
	body := `<div class="quote"><div class="quotetitle">someone wrote:</div><i>Can I do this?</i></div>Yes &amp; no.<br/><br/><br/><br/>See the  <a href="https://example.com">rulebook</a>,   page&nbsp;4.`
	want := "someone wrote:\nCan I do this?\nYes & no.\n\nSee the rulebook, page 4."
	if got := cleanPostBody(body); got != want {
		t.Errorf("cleanPostBody mismatch\n got: %q\nwant: %q", got, want)
	}
}
