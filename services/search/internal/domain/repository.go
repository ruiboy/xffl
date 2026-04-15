package domain

import "context"

// DocumentRepository persists and retrieves SearchDocuments.
type DocumentRepository interface {
	Index(ctx context.Context, doc SearchDocument) error
	Search(ctx context.Context, query SearchQuery) (SearchResult, error)
}
