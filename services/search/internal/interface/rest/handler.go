package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"xffl/services/search/internal/domain"
)

// Handler serves the Search REST API.
type Handler struct {
	repo domain.DocumentRepository
}

// NewHandler creates a Handler backed by the given repository.
func NewHandler(repo domain.DocumentRepository) *Handler {
	return &Handler{repo: repo}
}

// searchResponse is the JSON envelope for search results.
type searchResponse struct {
	Total     int              `json:"total"`
	Documents []documentJSON   `json:"documents"`
}

type documentJSON struct {
	ID     string         `json:"id"`
	Source string         `json:"source"`
	Type   string         `json:"type"`
	Data   map[string]any `json:"data"`
}

// ServeSearch handles GET /search?q=...&source=...&type=...
func (h *Handler) ServeSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := domain.SearchQuery{
		Q:      r.URL.Query().Get("q"),
		Source: r.URL.Query().Get("source"),
		Type:   r.URL.Query().Get("type"),
	}

	result, err := h.repo.Search(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "search failed", slog.Any("error", err))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	docs := make([]documentJSON, len(result.Documents))
	for i, d := range result.Documents {
		docs[i] = documentJSON{
			ID:     d.ID,
			Source: d.Source,
			Type:   d.Type,
			Data:   d.Data,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResponse{
		Total:     result.Total,
		Documents: docs,
	})
}
