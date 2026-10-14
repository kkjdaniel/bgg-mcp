package tools

import (
	"fmt"
	"net/url"
)

// BGGHomeURL is the page to credit when an answer uses BGG data without naming
// anything that has a page of its own.
const BGGHomeURL = "https://boardgamegeek.com"

// linkNote ends the description of every tool that returns BGG data. It sits
// in the description because clients pass tool descriptions to the model as
// part of the tool, while instructions inside a tool result are rightly
// treated as untrusted, and not every client passes on the server
// instructions.
const linkNote = ` Results give each game, user and forum thread a "link" to its BoardGameGeek page. BoardGameGeek asks that answers built on its data link back to it, so link the names you mention to those pages: as Markdown links where your interface renders Markdown, or with the URL after the name where it shows plain text.`

func withLinkNote(description string) string {
	return description + linkNote
}

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
