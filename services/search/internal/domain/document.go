package domain

// Source constants identify which service produced a document.
const (
	SourceAFL = "afl"
	SourceFFL = "ffl"
)

// Type constants identify the kind of document.
const (
	TypePlayerMatch  = "player_match"
	TypeFantasyScore = "fantasy_score"
)

// SearchDocument is the unit of data stored and retrieved from the search index.
type SearchDocument struct {
	ID     string         // "{source}_{type}_{id}"
	Source string         // SourceAFL or SourceFFL
	Type   string         // TypePlayerMatch or TypeFantasyScore
	Data   map[string]any // domain fields
}
