package application

import (
	"context"

	"xffl/services/search/internal/domain"
)

// IndexDocument indexes a single document into the search store.
type IndexDocument struct {
	repo domain.DocumentRepository
}

func NewIndexDocument(repo domain.DocumentRepository) *IndexDocument {
	return &IndexDocument{repo: repo}
}

func (uc *IndexDocument) Execute(ctx context.Context, doc domain.SearchDocument) error {
	return uc.repo.Index(ctx, doc)
}
