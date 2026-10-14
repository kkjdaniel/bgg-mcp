package tools

import (
	"fmt"
	"net/url"
)

// BGGHomeURL is the page to credit when an answer uses BGG data without naming
// anything that has a page of its own.
const BGGHomeURL = "https://boardgamegeek.com"

// BGGLink is something BGG has a page for, such as a designer or publisher.
type BGGLink struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// itemURL returns the BGG page of an item. itemType is the type the XML API
// gives an item or one of its links, such as boardgame, boardgameexpansion or
// boardgamedesigner, and falls back to boardgame when it is missing.
func itemURL(itemType string, id int) string {
	if itemType == "" {
		itemType = "boardgame"
	}
	return fmt.Sprintf("%s/%s/%d", BGGHomeURL, itemType, id)
}

func profileURL(username string) string {
	return BGGHomeURL + "/profile/" + url.PathEscape(username)
}

func collectionURL(username string) string {
	return BGGHomeURL + "/collection/user/" + url.PathEscape(username)
}

func threadURL(id int) string {
	return fmt.Sprintf("%s/thread/%d", BGGHomeURL, id)
}

func guildURL(id int) string {
	return fmt.Sprintf("%s/guild/%d", BGGHomeURL, id)
}
