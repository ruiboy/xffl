package forum

import "strings"

// teamByAuthor maps a forum username (lowercased) to the parser format for that
// poster's team. Populated for the current four teams; historical teams (Grand
// Pooh Bears, Buckleys, …) are added as their formats are met.
var teamByAuthor = map[string]string{
	"ruiboy":   "Ruiboys",
	"cheetahs": "Cheetahs",
	"slashers": "Slashers",
	"thc":      "THC",
}

// TeamForAuthor returns the parser format for a forum author, or "" if the author
// is unknown — the escape hatch: such a post is captured and flagged, not parsed,
// until its team/format is added.
func TeamForAuthor(author string) string {
	return teamByAuthor[strings.ToLower(strings.TrimSpace(author))]
}
