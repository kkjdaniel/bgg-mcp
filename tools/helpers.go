package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/search"
	"github.com/kkjdaniel/gogeek/v3/thing"
)

const selfUsernameMissingMessage = "BGG_USERNAME is not set. Either set it or provide your specific username instead of 'SELF'."

type EssentialGameInfo struct {
	ID                 int              `json:"id"`
	Name               string           `json:"name"`
	Link               string           `json:"link"`
	Description        string           `json:"description"`
	Year               int              `json:"year"`
	Complexity         float64          `json:"complexity"`
	Players            string           `json:"players"`
	BestPlayers        string           `json:"best_players,omitempty"`
	RecommendedPlayers string           `json:"recommended_players,omitempty"`
	PlayerCountVotes   int              `json:"player_count_votes,omitempty"`
	BGGRating          float64          `json:"bgg_rating"`
	BayesAverage       float64          `json:"bayes_average"`
	BGGRank            int              `json:"bgg_rank,omitempty"`
	LanguageDependence string           `json:"language_dependence,omitempty"`
	PlayTime           string           `json:"play_time"`
	MinAge             int              `json:"min_age"`
	Designers          []BGGLink        `json:"designers"`
	Publishers         []BGGLink        `json:"publishers"`
	Type               string           `json:"type"`
	Thumbnail          string           `json:"thumbnail"`
	Image              string           `json:"image"`
	Categories         []BGGLink        `json:"categories"`
	Mechanics          []BGGLink        `json:"mechanics"`
	Expands            []BGGLink        `json:"expands,omitempty"`
	Expansions         []BGGLink        `json:"expansions,omitempty"`
	NumRatings         int              `json:"num_ratings"`
	Owned              int              `json:"owned"`
	Wishing            int              `json:"wishing"`
	Trading            int              `json:"trading"`
	Wanting            int              `json:"wanting"`
	Videos             *GameVideos      `json:"videos,omitempty"`
	Versions           *GameVersions    `json:"versions,omitempty"`
	Comments           *GameComments    `json:"comments,omitempty"`
	Marketplace        *GameMarketplace `json:"marketplace,omitempty"`
}

// GameVersions holds the published editions of a game, such as each printing
// and localisation.
type GameVersions struct {
	Count    int           `json:"count"`
	Versions []GameVersion `json:"list"`
}

// GameVersion is one published edition of a game.
type GameVersion struct {
	Name       string   `json:"name"`
	Link       string   `json:"link"`
	Year       int      `json:"year,omitempty"`
	Publishers []string `json:"publishers,omitempty"`
	Languages  []string `json:"languages,omitempty"`
}

// GameComments holds one page of the comments users have left on a game, with
// the rating each gave it.
type GameComments struct {
	// TotalOnBGG counts every comment BGG holds for the game, across all pages.
	TotalOnBGG int           `json:"total_on_bgg"`
	Returned   int           `json:"returned"`
	Comments   []GameComment `json:"list"`
}

// GameComment is one user's comment on a game.
type GameComment struct {
	Username string  `json:"username"`
	Rating   float64 `json:"rating,omitempty"`
	Comment  string  `json:"comment"`
}

// GameMarketplace holds the copies of a game for sale on the BGG marketplace.
type GameMarketplace struct {
	Count    int                  `json:"count"`
	Listings []MarketplaceListing `json:"listings"`
}

// MarketplaceListing is one copy of a game for sale on the BGG marketplace.
type MarketplaceListing struct {
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Condition string  `json:"condition"`
	Notes     string  `json:"notes,omitempty"`
	Listed    string  `json:"listed"`
	Link      string  `json:"link"`
}

// GameVideos holds the community-submitted videos returned for a game, and the
// figure BGG publishes beside them.
type GameVideos struct {
	// TotalOnBGG is what BGG says it holds for the game. It is routinely far
	// larger than the number of entries below, so it counts the site's holdings
	// and never the videos in this response.
	TotalOnBGG int         `json:"total_on_bgg"`
	Returned   int         `json:"returned"`
	Videos     []GameVideo `json:"list"`
}

// GameVideo is one community-submitted video linked to a game.
type GameVideo struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Language string `json:"language"`
	Link     string `json:"link"`
	Username string `json:"username"`
	PostDate string `json:"post_date"`
}

func extractEssentialInfo(item thing.Item) EssentialGameInfo {
	info := EssentialGameInfo{
		ID:          item.ID,
		Name:        item.Name[0].Value,
		Link:        itemURL(item.Type, item.ID),
		Year:        item.YearPublished.Value,
		Description: item.Description,
		Type:        item.Type,
		Thumbnail:   item.Thumbnail,
		Image:       item.Image,
		MinAge:      item.MinAge.Value,
	}

	info.Videos = extractVideos(item)
	info.Versions = extractVersions(item)
	info.Comments = extractComments(item)
	info.Marketplace = extractMarketplace(item)
	info.LanguageDependence = languageDependence(item)

	if item.Statistics != nil && item.Statistics.AverageWeight.Value > 0 {
		info.Complexity = item.Statistics.AverageWeight.Value
	}

	info.Players = playerRange(item.MinPlayers.Value, item.MaxPlayers.Value)
	info.BestPlayers, info.RecommendedPlayers, info.PlayerCountVotes = suggestedPlayers(item)

	if item.Statistics != nil {
		if item.Statistics.Average.Value > 0 {
			info.BGGRating = item.Statistics.Average.Value
		}
		if item.Statistics.BayesAverage.Value > 0 {
			info.BayesAverage = item.Statistics.BayesAverage.Value
		}
		info.NumRatings = item.Statistics.UsersRated.Value
		for _, rank := range item.Statistics.Ranks {
			// The overall rank; the rest are family ranks such as strategy games.
			// An unranked game's value is "Not Ranked", which leaves it out.
			if rank.Name == "boardgame" {
				info.BGGRank, _ = strconv.Atoi(rank.Value)
			}
		}
		info.Owned = item.Statistics.Owned.Value
		info.Wishing = item.Statistics.Wishing.Value
		info.Trading = item.Statistics.Trading.Value
		info.Wanting = item.Statistics.Wanting.Value
	}

	info.PlayTime = playTimeRange(item.MinPlayTime.Value, item.MaxPlayTime.Value)

	for _, link := range item.Links {
		named := BGGLink{Name: link.Value, Link: itemURL(link.Type, link.ID)}
		switch link.Type {
		case "boardgamedesigner":
			info.Designers = append(info.Designers, named)
		case "boardgamepublisher":
			info.Publishers = append(info.Publishers, named)
		case "boardgamecategory":
			info.Categories = append(info.Categories, named)
		case "boardgamemechanic":
			info.Mechanics = append(info.Mechanics, named)
		case "boardgameexpansion":
			// On an expansion these point to the game it expands, and on a base
			// game to its expansions.
			if item.Type == "boardgameexpansion" {
				info.Expands = append(info.Expands, named)
			} else {
				info.Expansions = append(info.Expansions, named)
			}
		}
	}

	return info
}

// playerRange formats a player count such as "2-4", or "" when it is unknown.
func playerRange(min, max int) string {
	if min <= 0 || max <= 0 {
		return ""
	}
	if min == max {
		return fmt.Sprintf("%d", min)
	}
	return fmt.Sprintf("%d-%d", min, max)
}

// playTimeRange formats a play time such as "60-90 min", or "" when it is unknown.
func playTimeRange(min, max int) string {
	if min <= 0 || max <= 0 {
		return ""
	}
	if min == max {
		return fmt.Sprintf("%d min", min)
	}
	return fmt.Sprintf("%d-%d min", min, max)
}

// suggestedPlayers reads BGG's summary of the community poll on player counts,
// such as "Best with 3 players" and "Recommended with 2–4 players", down to the
// counts, along with how many people voted. A game nobody has voted on gives
// nothing.
func suggestedPlayers(item thing.Item) (best, recommended string, votes int) {
	for _, poll := range item.Polls {
		if poll.Name == "suggested_numplayers" {
			votes = poll.TotalVotes
		}
	}
	if votes == 0 {
		return "", "", 0
	}

	for _, summary := range item.PollSummaries {
		if summary.Name != "suggested_numplayers" {
			continue
		}
		for _, result := range summary.Results {
			switch {
			case strings.HasPrefix(result.Name, "best"):
				best = pollPlayerCounts(result.Value, "Best with")
			// BGG spells this one "recommmendedwith".
			case strings.HasPrefix(result.Name, "recomm"):
				recommended = pollPlayerCounts(result.Value, "Recommended with")
			}
		}
	}

	return best, recommended, votes
}

// pollPlayerCounts trims a poll summary such as "Recommended with 2–4 players"
// to its counts, written like the players field ("2-4").
func pollPlayerCounts(summary, prefix string) string {
	counts := strings.TrimPrefix(strings.TrimSpace(summary), prefix)
	counts = strings.TrimSuffix(counts, " players")
	counts = strings.TrimSuffix(counts, " player")
	return strings.ReplaceAll(strings.TrimSpace(counts), "–", "-")
}

// languageDependence is the answer most voted for in BGG's community poll on
// how much a game depends on its language, such as "No necessary in-game text".
func languageDependence(item thing.Item) string {
	answer, most := "", 0
	for _, poll := range item.Polls {
		if poll.Name != "language_dependence" {
			continue
		}
		for _, results := range poll.Results {
			for _, result := range results.Values {
				if result.NumVotes > most {
					answer, most = result.Value, result.NumVotes
				}
			}
		}
	}
	return answer
}

// extractVersions shapes the editions a query made WithVersions brings back.
// It gives nothing when there are none, which is also what a query made
// without the option looks like.
func extractVersions(item thing.Item) *GameVersions {
	if len(item.Versions) == 0 {
		return nil
	}
	versions := make([]GameVersion, 0, len(item.Versions))
	for _, v := range item.Versions {
		version := GameVersion{
			Link: itemURL(v.Type, v.ID),
			Year: v.YearPublished.Value,
		}
		if len(v.Name) > 0 {
			version.Name = v.Name[0].Value
		}
		for _, link := range v.Links {
			switch link.Type {
			case "boardgamepublisher":
				version.Publishers = append(version.Publishers, link.Value)
			case "language":
				version.Languages = append(version.Languages, link.Value)
			}
		}
		versions = append(versions, version)
	}
	return &GameVersions{Count: len(versions), Versions: versions}
}

// extractComments shapes the page of comments a query made WithComments brings
// back, and gives nothing for a query made without it.
func extractComments(item thing.Item) *GameComments {
	if item.Comments == nil {
		return nil
	}
	comments := make([]GameComment, 0, len(item.Comments.Comments))
	for _, c := range item.Comments.Comments {
		// A commenter who hasn't rated the game has the rating "N/A", which
		// leaves it out.
		rating, _ := strconv.ParseFloat(c.Rating, 64)
		comments = append(comments, GameComment{Username: c.Username, Rating: rating, Comment: c.Value})
	}
	return &GameComments{
		TotalOnBGG: item.Comments.Total,
		Returned:   len(comments),
		Comments:   comments,
	}
}

// listingConditions spells out the conditions BGG gives listings as single words.
var listingConditions = map[string]string{
	"likenew":  "like new",
	"verygood": "very good",
}

// extractMarketplace shapes the listings a query made WithMarketplace brings
// back. It gives nothing when there are none, which is also what a query made
// without the option looks like.
func extractMarketplace(item thing.Item) *GameMarketplace {
	if len(item.Marketplace) == 0 {
		return nil
	}
	listings := make([]MarketplaceListing, 0, len(item.Marketplace))
	for _, l := range item.Marketplace {
		condition := l.Condition.Value
		if spelled, ok := listingConditions[condition]; ok {
			condition = spelled
		}
		listed := l.ListDate.Value
		if t, err := time.Parse(time.RFC1123Z, listed); err == nil {
			listed = t.Format("2006-01-02")
		}
		listings = append(listings, MarketplaceListing{
			Price:     l.Price.Value,
			Currency:  l.Price.Currency,
			Condition: condition,
			Notes:     l.Notes.Value,
			Listed:    listed,
			Link:      l.Link.Href,
		})
	}
	return &GameMarketplace{Count: len(listings), Listings: listings}
}

// extractVideos shapes the videos a query made WithVideos brings back. A game
// queried without that option carries none, which is different from a game BGG
// holds no video for, so nothing is emitted at all in that case.
func extractVideos(item thing.Item) *GameVideos {
	if item.Videos == nil {
		return nil
	}
	list := make([]GameVideo, 0, len(item.Videos.Videos))
	for _, v := range item.Videos.Videos {
		list = append(list, GameVideo{
			Title:    v.Title,
			Category: v.Category,
			Language: v.Language,
			Link:     v.Link,
			Username: v.Username,
			PostDate: v.PostDate,
		})
	}
	return &GameVideos{
		TotalOnBGG: item.Videos.Total,
		Returned:   len(list),
		Videos:     list,
	}
}

// extractEssentialInfoList shapes several games. It leaves out each game's list
// of expansions, which can run to hundreds for a long-lived game, so a single
// game lookup is the way to get them.
func extractEssentialInfoList(items []thing.Item) []EssentialGameInfo {
	result := make([]EssentialGameInfo, len(items))
	for i, item := range items {
		result[i] = extractEssentialInfo(item)
		result[i].Expansions = nil
	}
	return result
}

func findBestGameMatch(ctx context.Context, client *gogeek.Client, gameName string) (*search.SearchResult, error) {
	searchResults, err := search.Query(ctx, client, gameName, search.WithExact())
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(searchResults.Items) == 0 {
		searchResults, err = search.Query(ctx, client, gameName)
		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}
		if len(searchResults.Items) == 0 {
			return nil, fmt.Errorf("no games found matching '%s'", gameName)
		}
	}

	bestMatch := &searchResults.Items[0]
	gameNameLower := strings.ToLower(gameName)

	for i := range searchResults.Items {
		item := &searchResults.Items[i]
		itemNameLower := strings.ToLower(item.Name.Value)

		if itemNameLower == gameNameLower {
			return item, nil
		}

		if bestMatch.Type == "boardgameexpansion" && item.Type == "boardgame" {
			if strings.Contains(itemNameLower, gameNameLower) || strings.Contains(gameNameLower, itemNameLower) {
				bestMatch = item
			}
		}

		if bestMatch.Type == item.Type {
			if strings.HasPrefix(itemNameLower, gameNameLower) && !strings.HasPrefix(strings.ToLower(bestMatch.Name.Value), gameNameLower) {
				bestMatch = item
			}
		}
	}

	return bestMatch, nil
}

// FetchGameInfo returns the essential information for a single game by its BGG ID.
func FetchGameInfo(ctx context.Context, client *gogeek.Client, gameID int) (*EssentialGameInfo, error) {
	things, err := thing.Query(ctx, client, []int{gameID})
	if err != nil {
		return nil, err
	}
	if len(things.Items) == 0 {
		return nil, fmt.Errorf("no game found with ID %d", gameID)
	}

	info := extractEssentialInfo(things.Items[0])
	return &info, nil
}
