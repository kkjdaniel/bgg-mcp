package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/search"
	"github.com/kkjdaniel/gogeek/v3/thing"
)

type EssentialGameInfo struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Year         int         `json:"year"`
	Complexity   float64     `json:"complexity"`
	Players      string      `json:"players"`
	BGGRating    float64     `json:"bgg_rating"`
	BayesAverage float64     `json:"bayes_average"`
	PlayTime     string      `json:"play_time"`
	MinAge       int         `json:"min_age"`
	Designer     string      `json:"designer"`
	Publisher    string      `json:"publisher"`
	Type         string      `json:"type"`
	Thumbnail    string      `json:"thumbnail"`
	Image        string      `json:"image"`
	Categories   []string    `json:"categories"`
	Mechanics    []string    `json:"mechanics"`
	NumRatings   int         `json:"num_ratings"`
	Owned        int         `json:"owned"`
	Wishing      int         `json:"wishing"`
	Trading      int         `json:"trading"`
	Wanting      int         `json:"wanting"`
	Videos       *GameVideos `json:"videos,omitempty"`
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
		Year:        item.YearPublished.Value,
		Description: item.Description,
		Type:        item.Type,
		Thumbnail:   item.Thumbnail,
		Image:       item.Image,
		MinAge:      item.MinAge.Value,
	}

	info.Videos = extractVideos(item)

	if item.Statistics != nil && item.Statistics.AverageWeight.Value > 0 {
		info.Complexity = item.Statistics.AverageWeight.Value
	}

	if item.MinPlayers.Value > 0 && item.MaxPlayers.Value > 0 {
		if item.MinPlayers.Value == item.MaxPlayers.Value {
			info.Players = fmt.Sprintf("%d", item.MinPlayers.Value)
		} else {
			info.Players = fmt.Sprintf("%d-%d", item.MinPlayers.Value, item.MaxPlayers.Value)
		}
	}

	if item.Statistics != nil {
		if item.Statistics.Average.Value > 0 {
			info.BGGRating = item.Statistics.Average.Value
		}
		if item.Statistics.BayesAverage.Value > 0 {
			info.BayesAverage = item.Statistics.BayesAverage.Value
		}
		info.NumRatings = item.Statistics.UsersRated.Value
		info.Owned = item.Statistics.Owned.Value
		info.Wishing = item.Statistics.Wishing.Value
		info.Trading = item.Statistics.Trading.Value
		info.Wanting = item.Statistics.Wanting.Value
	}

	if item.MinPlayTime.Value > 0 && item.MaxPlayTime.Value > 0 {
		if item.MinPlayTime.Value == item.MaxPlayTime.Value {
			info.PlayTime = fmt.Sprintf("%d min", item.MinPlayTime.Value)
		} else {
			info.PlayTime = fmt.Sprintf("%d-%d min", item.MinPlayTime.Value, item.MaxPlayTime.Value)
		}
	}

	var designers []string
	var publishers []string
	var categories []string
	var mechanics []string

	for _, link := range item.Links {
		switch link.Type {
		case "boardgamedesigner":
			designers = append(designers, link.Value)
		case "boardgamepublisher":
			publishers = append(publishers, link.Value)
		case "boardgamecategory":
			categories = append(categories, link.Value)
		case "boardgamemechanic":
			mechanics = append(mechanics, link.Value)
		}
	}

	if len(designers) > 0 {
		info.Designer = strings.Join(designers, ", ")
	}
	if len(publishers) > 0 {
		info.Publisher = strings.Join(publishers, ", ")
	}
	info.Categories = categories
	info.Mechanics = mechanics

	return info
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

func extractEssentialInfoList(items []thing.Item) []EssentialGameInfo {
	result := make([]EssentialGameInfo, len(items))
	for i, item := range items {
		result[i] = extractEssentialInfo(item)
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
