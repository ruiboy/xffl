package domain

// SearchQuery carries the parameters for a full-text search request.
type SearchQuery struct {
	Q      string // full-text search term
	Source string // optional: filter by SourceAFL or SourceFFL
	Type   string // optional: filter by TypePlayerMatch or TypeFantasyScore
}

// SearchResult is the response from a search query.
type SearchResult struct {
	Total     int
	Documents []SearchDocument
}
