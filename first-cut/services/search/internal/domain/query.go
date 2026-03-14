package domain

import "fmt"

// SearchQuery represents a search request
type SearchQuery struct {
	Query     string            `json:"query"`
	Filters   map[string]string `json:"filters,omitempty"`
	Source    string            `json:"source,omitempty"`    // "afl", "ffl", or empty for all
	Type      DocumentType      `json:"type,omitempty"`      // filter by document type
	Limit     int               `json:"limit,omitempty"`     // default 10
	Offset    int               `json:"offset,omitempty"`    // for pagination
	SortBy    string            `json:"sort_by,omitempty"`   // field to sort by
	SortOrder string            `json:"sort_order,omitempty"` // "asc" or "desc"
}

// SearchResult represents a single search result
type SearchResult struct {
	Document SearchDocument `json:"document"`
	Score    float64        `json:"score"`
	Snippet  string         `json:"snippet,omitempty"` // highlighted excerpt
}

// SearchResults represents the complete search response
type SearchResults struct {
	Results    []SearchResult `json:"results"`
	Total      int            `json:"total"`
	Query      string         `json:"query"`
	Took       int            `json:"took"` // milliseconds
	MaxScore   float64        `json:"max_score"`
	Pagination Pagination     `json:"pagination"`
}

// Pagination represents pagination information
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
	Pages  int `json:"pages"`
}

// NewSearchQuery creates a search query with defaults
func NewSearchQuery(query string) SearchQuery {
	return SearchQuery{
		Query:     query,
		Limit:     10,
		Offset:    0,
		SortOrder: "desc",
	}
}

// Validate checks if the search query is valid
func (q SearchQuery) Validate() error {
	if q.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if q.Limit < 0 || q.Limit > 100 {
		return fmt.Errorf("limit must be between 0 and 100")
	}
	if q.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	if q.SortOrder != "" && q.SortOrder != "asc" && q.SortOrder != "desc" {
		return fmt.Errorf("sort_order must be 'asc' or 'desc'")
	}
	return nil
}