package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kkjdaniel/gogeek/v3"
	"github.com/kkjdaniel/gogeek/v3/user"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// UserProfile is a BGG user's profile. The user, their buddies, their guilds and
// their top games each carry a link to their BGG page.
type UserProfile struct {
	ID               int          `json:"id"`
	Name             string       `json:"name"`
	Link             string       `json:"link"`
	CollectionLink   string       `json:"collection_link"`
	FirstName        string       `json:"first_name,omitempty"`
	LastName         string       `json:"last_name,omitempty"`
	AvatarLink       string       `json:"avatar_link,omitempty"`
	YearRegistered   int          `json:"year_registered,omitempty"`
	LastLogin        string       `json:"last_login,omitempty"`
	StateOrProvince  string       `json:"state_or_province,omitempty"`
	Country          string       `json:"country,omitempty"`
	WebAddress       string       `json:"web_address,omitempty"`
	XboxAccount      string       `json:"xbox_account,omitempty"`
	WiiAccount       string       `json:"wii_account,omitempty"`
	PSNAccount       string       `json:"psn_account,omitempty"`
	BattleNetAccount string       `json:"battlenet_account,omitempty"`
	SteamAccount     string       `json:"steam_account,omitempty"`
	TradeRating      int          `json:"trade_rating"`
	BuddiesTotal     int          `json:"buddies_total"`
	Buddies          []BGGLink    `json:"buddies"`
	GuildsTotal      int          `json:"guilds_total"`
	Guilds           []BGGLink    `json:"guilds"`
	TopGames         []RankedGame `json:"top_games"`
}

// RankedGame is a game in a user's top list.
type RankedGame struct {
	Rank int    `json:"rank"`
	ID   int    `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}

func newUserProfile(u *user.User) UserProfile {
	profile := UserProfile{
		ID:               u.ID,
		Name:             u.Name,
		Link:             profileURL(u.Name),
		CollectionLink:   collectionURL(u.Name),
		FirstName:        u.FirstName.Value,
		LastName:         u.LastName.Value,
		AvatarLink:       u.AvatarLink.Value,
		YearRegistered:   u.YearRegistered.Value,
		LastLogin:        u.LastLogin.Value,
		StateOrProvince:  u.StateOrProvince.Value,
		Country:          u.Country.Value,
		WebAddress:       u.WebAddress.Value,
		XboxAccount:      u.XboxAccount.Value,
		WiiAccount:       u.WiiAccount.Value,
		PSNAccount:       u.PSNAccount.Value,
		BattleNetAccount: u.BattleNetAccount.Value,
		SteamAccount:     u.SteamAccount.Value,
		TradeRating:      u.TradeRating.Value,
		BuddiesTotal:     u.Buddies.Total,
		Buddies:          make([]BGGLink, 0, len(u.Buddies.Buddy)),
		GuildsTotal:      u.Guilds.Total,
		Guilds:           make([]BGGLink, 0, len(u.Guilds.Guild)),
		TopGames:         make([]RankedGame, 0, len(u.Top.Items)),
	}

	for _, buddy := range u.Buddies.Buddy {
		profile.Buddies = append(profile.Buddies, BGGLink{Name: buddy.Name, Link: profileURL(buddy.Name)})
	}
	for _, guild := range u.Guilds.Guild {
		profile.Guilds = append(profile.Guilds, BGGLink{Name: guild.Name, Link: guildURL(guild.ID)})
	}
	// Top list items have the type "thing", which is not a page on BGG, so the
	// list's domain (such as boardgame) is used instead
	for _, item := range u.Top.Items {
		profile.TopGames = append(profile.TopGames, RankedGame{
			Rank: item.Rank,
			ID:   item.ID,
			Name: item.Name,
			Link: itemURL(u.Top.Domain, item.ID),
		})
	}

	return profile
}

func UserTool(client *gogeek.Client, selfUsername string) (mcp.Tool, server.ToolHandlerFunc) {
	tool := mcp.NewTool("bgg-user",
		mcp.WithDescription(withLinkNote("Find details about a specific user on BoardGameGeek (BGG)")),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the BoardGameGeek (BGG) user. When the user refers to themselves (me, my, I), use 'SELF' as the value."),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.GetArguments()
		name := arguments["username"].(string)

		if name == "SELF" {
			if selfUsername == "" {
				return mcp.NewToolResultText(selfUsernameMissingMessage), nil
			}
			name = selfUsername
		}

		userDetails, err := user.Query(ctx, client, name)
		if err != nil {
			return mcp.NewToolResultText(err.Error()), nil
		}

		out, err := json.Marshal(newUserProfile(userDetails))
		if err != nil {
			return mcp.NewToolResultText(fmt.Sprintf("Error formatting results: %v", err)), nil
		}
		return mcp.NewToolResultText(string(out)), nil

	}

	return tool, handler
}
