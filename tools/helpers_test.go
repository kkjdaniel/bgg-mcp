package tools

import (
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/kkjdaniel/gogeek/v3/thing"
)

// thingXML is the shape of a game from the thing endpoint, with the polls,
// ranks and links BGG sends for every game, and the versions, comments and
// marketplace listings that only come with their options.
const thingXML = `<items>
  <item type="boardgame" id="13">
    <name type="primary" sortindex="1" value="Catan"/>
    <yearpublished value="1995"/>
    <minplayers value="3"/>
    <maxplayers value="4"/>
    <poll name="suggested_numplayers" title="User Suggested Number of Players" totalvotes="60">
      <results numplayers="3"><result value="Best" numvotes="25"/></results>
    </poll>
    <poll-summary name="suggested_numplayers" title="User Suggested Number of Players">
      <result name="bestwith" value="Best with 4 players"/>
      <result name="recommmendedwith" value="Recommended with 3–4 players"/>
    </poll-summary>
    <poll name="language_dependence" title="Language Dependence" totalvotes="9">
      <results>
        <result level="1" value="No necessary in-game text" numvotes="2"/>
        <result level="2" value="Some necessary text - easily memorized or small crib sheet" numvotes="7"/>
      </results>
    </poll>
    <link type="boardgameexpansion" id="325" value="Catan: Seafarers"/>
    <link type="boardgameexpansion" id="926" value="Catan: Cities &amp; Knights"/>
    <statistics page="1">
      <ratings>
        <usersrated value="120000"/>
        <average value="7.1"/>
        <bayesaverage value="6.9"/>
        <ranks>
          <rank type="subtype" id="1" name="boardgame" friendlyname="Board Game Rank" value="567" bayesaverage="6.9"/>
          <rank type="family" id="5499" name="familygames" friendlyname="Family Game Rank" value="120" bayesaverage="6.9"/>
        </ranks>
      </ratings>
    </statistics>
    <versions>
      <item type="boardgameversion" id="701">
        <link type="boardgameversion" id="13" value="Catan" inbound="true"/>
        <name type="primary" sortindex="1" value="German first edition"/>
        <yearpublished value="1995"/>
        <link type="boardgamepublisher" id="37" value="KOSMOS"/>
        <link type="language" id="2188" value="German"/>
      </item>
    </versions>
    <comments page="1" totalitems="8543">
      <comment username="player_one" rating="8" value="A modern classic."/>
      <comment username="player_two" rating="N/A" value="Traded away."/>
    </comments>
    <marketplacelistings>
      <listing>
        <listdate value="Tue, 05 Mar 2024 10:00:00 +0000"/>
        <price currency="USD" value="45.00"/>
        <condition value="likenew"/>
        <notes value="Played once."/>
        <link href="https://boardgamegeek.com/market/product/101" title="marketlisting"/>
      </listing>
    </marketplacelistings>
  </item>
  <item type="boardgameexpansion" id="325">
    <name type="primary" sortindex="1" value="Catan: Seafarers"/>
    <poll name="suggested_numplayers" title="User Suggested Number of Players" totalvotes="0"/>
    <poll-summary name="suggested_numplayers" title="User Suggested Number of Players">
      <result name="bestwith" value="Best with 4 players"/>
    </poll-summary>
    <link type="boardgameexpansion" id="13" value="Catan"/>
  </item>
</items>`

func parseThings(t *testing.T) []thing.Item {
	t.Helper()
	var items thing.Items
	if err := xml.Unmarshal([]byte(thingXML), &items); err != nil {
		t.Fatal(err)
	}
	return items.Items
}

func TestExtractEssentialInfoCommunityData(t *testing.T) {
	items := parseThings(t)
	info := extractEssentialInfo(items[0])

	if info.BestPlayers != "4" || info.RecommendedPlayers != "3-4" || info.PlayerCountVotes != 60 {
		t.Errorf("player poll = best %q, recommended %q, votes %d; want \"4\", \"3-4\", 60",
			info.BestPlayers, info.RecommendedPlayers, info.PlayerCountVotes)
	}
	if info.BGGRank != 567 {
		t.Errorf("rank = %d, want the overall board game rank 567", info.BGGRank)
	}
	if want := "Some necessary text - easily memorized or small crib sheet"; info.LanguageDependence != want {
		t.Errorf("language dependence = %q, want the most voted answer %q", info.LanguageDependence, want)
	}

	wantExpansions := []BGGLink{
		{"Catan: Seafarers", "https://boardgamegeek.com/boardgameexpansion/325"},
		{"Catan: Cities & Knights", "https://boardgamegeek.com/boardgameexpansion/926"},
	}
	if !reflect.DeepEqual(info.Expansions, wantExpansions) || info.Expands != nil {
		t.Errorf("base game expansions = %+v, expands = %+v; want %+v and none", info.Expansions, info.Expands, wantExpansions)
	}

	expansion := extractEssentialInfo(items[1])
	if len(expansion.Expands) != 1 || expansion.Expands[0].Name != "Catan" || expansion.Expansions != nil {
		t.Errorf("an expansion should point to the game it expands, got expands %+v, expansions %+v", expansion.Expands, expansion.Expansions)
	}
	if expansion.BestPlayers != "" || expansion.PlayerCountVotes != 0 {
		t.Errorf("a game nobody voted on should leave the player poll out, got %q (%d votes)", expansion.BestPlayers, expansion.PlayerCountVotes)
	}
}

func TestExtractEssentialInfoListDropsExpansionLists(t *testing.T) {
	infos := extractEssentialInfoList(parseThings(t))
	if infos[0].Expansions != nil {
		t.Errorf("a multi-game result should leave out each game's expansions, got %+v", infos[0].Expansions)
	}
	if len(infos[1].Expands) != 1 {
		t.Errorf("a multi-game result should keep the game an expansion expands, got %+v", infos[1].Expands)
	}
}

func TestExtractEssentialInfoOptionalData(t *testing.T) {
	items := parseThings(t)
	info := extractEssentialInfo(items[0])

	wantVersions := &GameVersions{Count: 1, Versions: []GameVersion{{
		Name:       "German first edition",
		Link:       "https://boardgamegeek.com/boardgameversion/701",
		Year:       1995,
		Publishers: []string{"KOSMOS"},
		Languages:  []string{"German"},
	}}}
	if !reflect.DeepEqual(info.Versions, wantVersions) {
		t.Errorf("versions = %+v, want %+v", info.Versions, wantVersions)
	}

	wantComments := &GameComments{TotalOnBGG: 8543, Returned: 2, Comments: []GameComment{
		{Username: "player_one", Rating: 8, Comment: "A modern classic."},
		{Username: "player_two", Comment: "Traded away."},
	}}
	if !reflect.DeepEqual(info.Comments, wantComments) {
		t.Errorf("comments = %+v, want %+v", info.Comments, wantComments)
	}

	wantMarketplace := &GameMarketplace{Count: 1, Listings: []MarketplaceListing{{
		Price:     45,
		Currency:  "USD",
		Condition: "like new",
		Notes:     "Played once.",
		Listed:    "2024-03-05",
		Link:      "https://boardgamegeek.com/market/product/101",
	}}}
	if !reflect.DeepEqual(info.Marketplace, wantMarketplace) {
		t.Errorf("marketplace = %+v, want %+v", info.Marketplace, wantMarketplace)
	}

	plain := extractEssentialInfo(items[1])
	if plain.Versions != nil || plain.Comments != nil || plain.Marketplace != nil {
		t.Errorf("a game queried without the options should carry none of their fields, got %+v", plain)
	}
}
