package tools

import (
	"encoding/xml"
	"reflect"
	"testing"

	"github.com/kkjdaniel/gogeek/v3/collection"
)

// statsCollectionXML is the shape of a collection queried with stats=1.
const statsCollectionXML = `<items totalitems="2">
  <item objecttype="thing" objectid="342942" subtype="boardgame" collid="1">
    <name sortindex="1">Ark Nova</name>
    <yearpublished>2021</yearpublished>
    <image>https://example.com/full.jpg</image>
    <thumbnail>https://example.com/thumb.jpg</thumbnail>
    <stats minplayers="1" maxplayers="4" minplaytime="90" maxplaytime="150" playingtime="150" numowned="90000">
      <rating value="9.5">
        <usersrated value="60000"/>
        <average value="8.53"/>
        <bayesaverage value="8.36"/>
        <stddev value="1.4"/>
        <median value="0"/>
        <ranks>
          <rank type="subtype" id="1" name="boardgame" friendlyname="Board Game Rank" value="4" bayesaverage="8.36"/>
          <rank type="family" id="5497" name="strategygames" friendlyname="Strategy Game Rank" value="3" bayesaverage="8.4"/>
        </ranks>
      </rating>
    </stats>
    <status own="1" prevowned="0" fortrade="1" want="0" wanttoplay="1" wanttobuy="0" wishlist="0" preordered="0" lastmodified="2025-01-01 00:00:00"/>
    <numplays>12</numplays>
    <comment>Best in 2-player.</comment>
  </item>
  <item objecttype="thing" objectid="999" subtype="boardgame" collid="2">
    <name sortindex="1">Unrated Promo</name>
    <stats minplayers="2" maxplayers="2" minplaytime="30" maxplaytime="30" playingtime="30" numowned="10">
      <rating value="N/A">
        <average value="0"/>
        <ranks>
          <rank type="subtype" id="1" name="boardgame" friendlyname="Board Game Rank" value="Not Ranked" bayesaverage="Not Ranked"/>
        </ranks>
      </rating>
    </stats>
    <status own="1" prevowned="0" fortrade="0" want="0" wanttoplay="0" wanttobuy="0" wishlist="0" preordered="0" lastmodified="2025-01-01 00:00:00"/>
    <numplays>0</numplays>
  </item>
</items>`

func TestNewCollectionResultStats(t *testing.T) {
	var c collection.Collection
	if err := xml.Unmarshal([]byte(statsCollectionXML), &c); err != nil {
		t.Fatal(err)
	}

	result := NewCollectionResult("someone", c.Items)
	if result.Count != 2 {
		t.Errorf("count = %d, want 2", result.Count)
	}

	want := CollectionGame{
		ID:             342942,
		Name:           "Ark Nova",
		Link:           "https://boardgamegeek.com/boardgame/342942",
		Year:           2021,
		PersonalRating: 9.5,
		BGGRating:      8.53,
		BGGRank:        4,
		Players:        "1-4",
		PlayTime:       "90-150 min",
		NumPlays:       12,
		Status:         []string{"owned", "for trade", "want to play"},
		Comment:        "Best in 2-player.",
	}
	if got := result.Items[0]; !reflect.DeepEqual(got, want) {
		t.Errorf("rated game = %+v\nwant %+v", got, want)
	}

	unrated := result.Items[1]
	if unrated.PersonalRating != 0 || unrated.BGGRating != 0 || unrated.BGGRank != 0 {
		t.Errorf("an unrated, unranked game should leave its rating and rank out, got %+v", unrated)
	}
	if unrated.Players != "2" || unrated.PlayTime != "30 min" {
		t.Errorf("fixed player count and play time = %q, %q, want \"2\", \"30 min\"", unrated.Players, unrated.PlayTime)
	}
}

func TestBuildCollectionOptionsRequestsStats(t *testing.T) {
	options := buildCollectionOptions(map[string]interface{}{"username": "someone"})
	if len(options) == 0 {
		t.Fatal("expected options")
	}

	// Run the options against a collection query's parameters to see what
	// they set, since collection.Option is opaque.
	params := map[string]string{}
	for _, opt := range options {
		values := make(map[string][]string)
		if err := opt(values); err != nil {
			t.Fatal(err)
		}
		for k, v := range values {
			params[k] = v[0]
		}
	}
	if params["stats"] != "1" {
		t.Errorf("collection queries should ask for stats, got params %v", params)
	}
	if params["own"] != "1" {
		t.Errorf("collection queries should default to owned games, got params %v", params)
	}
}
