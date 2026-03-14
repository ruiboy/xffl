package services

import (
	"context"
	"xffl/services/search/internal/domain"
)

// searchRepository defines the interface for search operations
type searchRepository interface {
	Index(ctx context.Context, doc domain.SearchDocument) error
	Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResults, error)
	Delete(ctx context.Context, documentID string) error
	BulkIndex(ctx context.Context, docs []domain.SearchDocument) error
	HealthCheck(ctx context.Context) error
}

// SearchService handles search operations
type SearchService struct {
	searchRepo searchRepository
}

// NewSearchService creates a new SearchService
func NewSearchService(searchRepo searchRepository) *SearchService {
	return &SearchService{
		searchRepo: searchRepo,
	}
}

// Search performs a search query
func (s *SearchService) Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResults, error) {
	if err := query.Validate(); err != nil {
		return domain.SearchResults{}, err
	}

	return s.searchRepo.Search(ctx, query)
}

// IndexDocument indexes a single document
func (s *SearchService) IndexDocument(ctx context.Context, doc domain.SearchDocument) error {
	return s.searchRepo.Index(ctx, doc)
}

// DeleteDocument removes a document from the index
func (s *SearchService) DeleteDocument(ctx context.Context, documentID string) error {
	return s.searchRepo.Delete(ctx, documentID)
}

// BulkIndexDocuments indexes multiple documents efficiently
func (s *SearchService) BulkIndexDocuments(ctx context.Context, docs []domain.SearchDocument) error {
	if len(docs) == 0 {
		return nil
	}
	return s.searchRepo.BulkIndex(ctx, docs)
}

// HealthCheck verifies the search engine is available
func (s *SearchService) HealthCheck(ctx context.Context) error {
	return s.searchRepo.HealthCheck(ctx)
}