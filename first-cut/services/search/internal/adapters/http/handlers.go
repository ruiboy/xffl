package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"xffl/services/search/internal/domain"
	"xffl/services/search/internal/services"
)

// SearchHandler handles HTTP requests for search operations
type SearchHandler struct {
	searchService *services.SearchService
}

// NewSearchHandler creates a new search HTTP handler
func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// HandleSearch processes search requests
func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	searchQuery := domain.NewSearchQuery(query)
	
	// Parse optional parameters
	if source := r.URL.Query().Get("source"); source != "" {
		searchQuery.Source = source
	}
	
	if docType := r.URL.Query().Get("type"); docType != "" {
		searchQuery.Type = domain.DocumentType(docType)
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			searchQuery.Limit = limit
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			searchQuery.Offset = offset
		}
	}
	
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		searchQuery.SortBy = sortBy
	}
	
	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		searchQuery.SortOrder = strings.ToLower(sortOrder)
	}

	// Execute search
	results, err := h.searchService.Search(r.Context(), searchQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// HandleHealthCheck checks if the search service is healthy
func (h *SearchHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.searchService.HealthCheck(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// IndexHandler handles manual document indexing (for admin/testing)
type IndexHandler struct {
	searchService *services.SearchService
}

// NewIndexHandler creates a new index HTTP handler
func NewIndexHandler(searchService *services.SearchService) *IndexHandler {
	return &IndexHandler{
		searchService: searchService,
	}
}

// HandleIndexDocument handles manual document indexing
func (h *IndexHandler) HandleIndexDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var doc domain.SearchDocument
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.searchService.IndexDocument(r.Context(), doc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"document_id": doc.ID,
	})
}

// HandleDeleteDocument handles manual document deletion
func (h *IndexHandler) HandleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract document ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/admin/index/")
	if path == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}

	if err := h.searchService.DeleteDocument(r.Context(), path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"document_id": path,
	})
}