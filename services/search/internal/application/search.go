package application

import (
	"context"

	"xffl/services/search/internal/domain"
)

// Search executes a full-text search query.
type Search struct {
	repo domain.DocumentRepository
}

func NewSearch(repo domain.DocumentRepository) *Search {
	return &Search{repo: repo}
}

func (uc *Search) Execute(ctx context.Context, query domain.SearchQuery) (domain.SearchResult, error) {
	return uc.repo.Search(ctx, query)
}
