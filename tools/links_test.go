package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kkjdaniel/gogeek/v3/collection"
	"github.com/kkjdaniel/gogeek/v3/thing"
	"github.com/kkjdaniel/gogeek/v3/user"
)

func TestExtractEssentialInfoLinks(t *testing.T) {
	item := thing.Item{
		Type: "boardgameexpansion",
		ID:   325,
		Name: []thing.Name{{Value: "Catan: Seafarers"}},
		Links: []thing.Link{
			{Type: "boardgamedesigner", ID: 11, Value: "Klaus Teuber"},
			{Type: "boardgamepublisher", ID: 37, Value: "KOSMOS"},
			{Type: "boardgamecategory", ID: 1026, Value: "Negotiation"},
			{Type: "boardgamemechanic", ID: 2072, Value: "Dice Rolling"},
			{Type: "boardgameartist", ID: 12, Value: "An Artist"},
		},
	}

	info := extractEssentialInfo(item)

	if want := "https://boardgamegeek.com/boardgameexpansion/325"; info.Link != want {
		t.Errorf("game link = %q, want %q", info.Link, want)
	}

	cases := []struct {
		field string
		got   []BGGLink
		want  BGGLink
	}{
		{"designers", info.Designers, BGGLink{"Klaus Teuber", "https://boardgamegeek.com/boardgamedesigner/11"}},
		{"publishers", info.Publishers, BGGLink{"KOSMOS", "https://boardgamegeek.com/boardgamepublisher/37"}},
		{"categories", info.Categories, BGGLink{"Negotiation", "https://boardgamegeek.com/boardgamecategory/1026"}},
		{"mechanics", info.Mechanics, BGGLink{"Dice Rolling", "https://boardgamegeek.com/boardgamemechanic/2072"}},
	}
	for _, c := range cases {
		if len(c.got) != 1 || c.got[0] != c.want {
			t.Errorf("%s = %+v, want [%+v]", c.field, c.got, c.want)
		}
	}
}

func TestNewCollectionResultLinks(t *testing.T) {
	result := NewCollectionResult("some user", []collection.CollectionItem{
		{ObjectID: 13, Subtype: "boardgame", Name: "Catan"},
		{ObjectID: 325, Subtype: "boardgameexpansion", Name: "Catan: Seafarers"},
	})

	if want := "https://boardgamegeek.com/profile/some%20user"; result.UserLink != want {
		t.Errorf("user link = %q, want %q", result.UserLink, want)
	}
	if want := "https://boardgamegeek.com/collection/user/some%20user"; result.CollectionLink != want {
		t.Errorf("collection link = %q, want %q", result.CollectionLink, want)
	}

	wantLinks := []string{
		"https://boardgamegeek.com/boardgame/13",
		"https://boardgamegeek.com/boardgameexpansion/325",
	}
	for i, want := range wantLinks {
		if got := result.Items[i].Link; got != want {
			t.Errorf("item %d link = %q, want %q", i, got, want)
		}
	}

	out, err := json.Marshal(result.Items[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"name":"Catan"`) || !strings.Contains(string(out), `"link":"https://boardgamegeek.com/boardgame/13"`) {
		t.Errorf("collection item should carry its name and link, got %s", out)
	}

	empty, err := json.Marshal(NewCollectionResult("someone", nil))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(empty), `"items":[]`) {
		t.Errorf("an empty collection should have an empty items list, got %s", empty)
	}
}

func TestAnalyseTradeOpportunitiesLinks(t *testing.T) {
	owned := &collection.Collection{Items: []collection.CollectionItem{
		{ObjectID: 13, Subtype: "boardgame", Name: "Catan"},
	}}
	wishlist := &collection.Collection{Items: []collection.CollectionItem{
		{ObjectID: 13, Subtype: "boardgame", Name: "Catan"},
	}}

	result := analyseTradeOpportunities("alice", "bob", owned, wishlist)

	if result.User1Link != "https://boardgamegeek.com/profile/alice" || result.User2Link != "https://boardgamegeek.com/profile/bob" {
		t.Errorf("user links = %q, %q", result.User1Link, result.User2Link)
	}
	for _, items := range [][]TradeItem{result.User1HasWanted, result.User2Wishlist} {
		if len(items) != 1 || items[0].Link != "https://boardgamegeek.com/boardgame/13" {
			t.Errorf("trade items = %+v, want one linked to Catan's page", items)
		}
	}
}

func TestNewUserProfileLinks(t *testing.T) {
	u := &user.User{
		ID:   1,
		Name: "alice",
		Buddies: user.Buddies{Total: 1, Buddy: []user.Buddy{
			{ID: 2, Name: "bob"},
		}},
		Guilds: user.Guilds{Total: 1, Guild: []user.Guild{
			{ID: 2001, Name: "A Guild"},
		}},
		Top: user.Top{Domain: "boardgame", Items: []user.TopItem{
			{Rank: 1, Type: "thing", ID: 224517, Name: "Brass: Birmingham"},
		}},
	}

	profile := newUserProfile(u)

	if profile.Link != "https://boardgamegeek.com/profile/alice" {
		t.Errorf("profile link = %q", profile.Link)
	}
	if profile.CollectionLink != "https://boardgamegeek.com/collection/user/alice" {
		t.Errorf("collection link = %q", profile.CollectionLink)
	}
	if want := (BGGLink{"bob", "https://boardgamegeek.com/profile/bob"}); len(profile.Buddies) != 1 || profile.Buddies[0] != want {
		t.Errorf("buddies = %+v, want [%+v]", profile.Buddies, want)
	}
	if want := (BGGLink{"A Guild", "https://boardgamegeek.com/guild/2001"}); len(profile.Guilds) != 1 || profile.Guilds[0] != want {
		t.Errorf("guilds = %+v, want [%+v]", profile.Guilds, want)
	}
	if len(profile.TopGames) != 1 || profile.TopGames[0].Link != "https://boardgamegeek.com/boardgame/224517" {
		t.Errorf("top games = %+v, want Brass: Birmingham linked to its boardgame page", profile.TopGames)
	}
}
