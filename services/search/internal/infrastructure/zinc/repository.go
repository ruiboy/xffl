package zinc

import (
	"context"
	"fmt"

	"xffl/services/search/internal/domain"
)

// Repository implements domain.DocumentRepository using ZincSearch.
type Repository struct {
	client *Client
}

// NewRepository wraps a Client as a domain.DocumentRepository.
func NewRepository(client *Client) *Repository {
	return &Repository{client: client}
}

// EnsureIndex delegates to the underlying client.
func (r *Repository) EnsureIndex(ctx context.Context) error {
	return r.client.EnsureIndex(ctx)
}

// Index stores a SearchDocument in Zinc. source and type are stored at the
// top level alongside the domain data fields so they are filterable.
func (r *Repository) Index(ctx context.Context, doc domain.SearchDocument) error {
	body := make(map[string]any, len(doc.Data)+2)
	for k, v := range doc.Data {
		body[k] = v
	}
	body["source"] = doc.Source
	body["type"] = doc.Type

	return r.client.indexDoc(ctx, doc.ID, body)
}

// Search runs a full-text query against Zinc and post-filters by source/type.
//
// ZincSearch v0.4 does not support field-specific queries in its ES-compatible
// _search endpoint — all term/match/query_string queries are treated as
// all-fields searches. We therefore apply source/type filtering in Go after
// fetching results from Zinc.
func (r *Repository) Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResult, error) {
	resp, err := r.client.search(ctx, buildQuery(query))
	if err != nil {
		return domain.SearchResult{}, fmt.Errorf("zinc repository search: %w", err)
	}

	// Reserved fields stored by Zinc that are not part of the domain document.
	reserved := map[string]bool{"source": true, "type": true, "@timestamp": true}

	var docs []domain.SearchDocument
	for _, hit := range resp.Hits.Hits {
		source, _ := hit.Source["source"].(string)
		typ, _ := hit.Source["type"].(string)

		if query.Source != "" && source != query.Source {
			continue
		}
		if query.Type != "" && typ != query.Type {
			continue
		}

		data := make(map[string]any, len(hit.Source))
		for k, v := range hit.Source {
			if !reserved[k] {
				data[k] = v
			}
		}

		docs = append(docs, domain.SearchDocument{
			ID:     hit.ID,
			Source: source,
			Type:   typ,
			Data:   data,
		})
	}

	return domain.SearchResult{
		Total:     len(docs),
		Documents: docs,
	}, nil
}

// buildQuery returns the Zinc query DSL body. Source/type filtering is applied
// in Go (post-fetch) because ZincSearch v0.4 ignores field-specific clauses.
func buildQuery(q domain.SearchQuery) map[string]any {
	var query any
	if q.Q != "" {
		query = map[string]any{
			"query_string": map[string]any{"query": q.Q},
		}
	} else {
		query = map[string]any{"match_all": map[string]any{}}
	}

	return map[string]any{
		"query": query,
		"size":  1000,
	}
}
