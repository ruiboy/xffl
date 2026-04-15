package typesense

import (
	"context"
	"fmt"
	"strings"

	"xffl/services/search/internal/domain"
)

// Repository implements domain.DocumentRepository using Typesense.
type Repository struct {
	client *Client
}

// NewRepository wraps a Client as a domain.DocumentRepository.
func NewRepository(client *Client) *Repository {
	return &Repository{client: client}
}

// EnsureCollection delegates to the underlying client.
func (r *Repository) EnsureCollection(ctx context.Context) error {
	return r.client.EnsureCollection(ctx)
}

// Index stores a SearchDocument in Typesense via upsert. source and type are
// stored at the top level alongside the domain data fields.
func (r *Repository) Index(ctx context.Context, doc domain.SearchDocument) error {
	body := make(map[string]any, len(doc.Data)+3)
	for k, v := range doc.Data {
		body[k] = v
	}
	body["id"] = doc.ID
	body["source"] = doc.Source
	body["type"] = doc.Type

	return r.client.upsertDoc(ctx, body)
}

// Search runs a query against Typesense and maps results to domain types.
// Filtering by source and type uses Typesense's native filter_by, so results
// are correct without post-filtering.
func (r *Repository) Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResult, error) {
	q := "*"
	if query.Q != "" {
		q = query.Q
	}

	resp, err := r.client.search(ctx, q, "source,type", buildFilterBy(query))
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("typesense repository search: %w", err)
	}

	reserved := map[string]bool{"id": true, "source": true, "type": true}

	docs := make([]domain.SearchDocument, 0, len(resp.Hits))
	for _, hit := range resp.Hits {
		id, _ := hit.Document["id"].(string)
		source, _ := hit.Document["source"].(string)
		typ, _ := hit.Document["type"].(string)

		data := make(map[string]any, len(hit.Document))
		for k, v := range hit.Document {
			if !reserved[k] {
				data[k] = v
			}
		}

		docs = append(docs, domain.SearchDocument{
			ID:     id,
			Source: source,
			Type:   typ,
			Data:   data,
		})
	}

	return domain.SearchResult{
		Total:     resp.Found,
		Documents: docs,
	}, nil
}

// buildFilterBy constructs the Typesense filter_by parameter from source/type.
func buildFilterBy(q domain.SearchQuery) string {
	var parts []string
	if q.Source != "" {
		parts = append(parts, fmt.Sprintf("source:=%s", q.Source))
	}
	if q.Type != "" {
		parts = append(parts, fmt.Sprintf("type:=%s", q.Type))
	}
	return strings.Join(parts, " && ")
}
