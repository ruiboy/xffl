package zinc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"xffl/services/search/internal/domain"
)

// ZincRepository implements searchRepository interface for Zinc search engine
type ZincRepository struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
	indexName  string
}

// ZincConfig holds configuration for Zinc repository
type ZincConfig struct {
	BaseURL   string
	Username  string
	Password  string
	IndexName string
	Timeout   time.Duration
}

// NewZincRepository creates a new Zinc search repository
func NewZincRepository(config ZincConfig) *ZincRepository {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.IndexName == "" {
		config.IndexName = "xffl"
	}
	if config.Username == "" {
		config.Username = "admin"
	}
	if config.Password == "" {
		config.Password = "admin"
	}

	return &ZincRepository{
		baseURL:   strings.TrimSuffix(config.BaseURL, "/"),
		indexName: config.IndexName,
		username:  config.Username,
		password:  config.Password,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Index indexes a single document in Zinc
func (z *ZincRepository) Index(ctx context.Context, doc domain.SearchDocument) error {
	url := fmt.Sprintf("%s/api/%s/_doc/%s", z.baseURL, z.indexName, doc.ID)
	
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(z.username, z.password)

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("zinc index request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Search performs a search query against Zinc
func (z *ZincRepository) Search(ctx context.Context, query domain.SearchQuery) (domain.SearchResults, error) {
	url := fmt.Sprintf("%s/api/%s/_search", z.baseURL, z.indexName)
	
	// Build Zinc search request
	searchReq := z.buildZincQuery(query)
	
	jsonData, err := json.Marshal(searchReq)
	if err != nil {
		return domain.SearchResults{}, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return domain.SearchResults{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(z.username, z.password)

	start := time.Now()
	resp, err := z.httpClient.Do(req)
	if err != nil {
		return domain.SearchResults{}, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return domain.SearchResults{}, fmt.Errorf("zinc search request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var zincResp ZincSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&zincResp); err != nil {
		return domain.SearchResults{}, fmt.Errorf("failed to decode search response: %w", err)
	}

	took := int(time.Since(start).Milliseconds())
	return z.convertZincResponse(zincResp, query, took), nil
}

// Delete removes a document from the Zinc index
func (z *ZincRepository) Delete(ctx context.Context, documentID string) error {
	url := fmt.Sprintf("%s/api/%s/_doc/%s", z.baseURL, z.indexName, documentID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.SetBasicAuth(z.username, z.password)

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("zinc delete request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// BulkIndex indexes multiple documents efficiently
func (z *ZincRepository) BulkIndex(ctx context.Context, docs []domain.SearchDocument) error {
	url := fmt.Sprintf("%s/api/_bulk", z.baseURL)
	
	var bulkBody strings.Builder
	for _, doc := range docs {
		// Action line
		action := map[string]interface{}{
			"index": map[string]string{
				"_index": z.indexName,
				"_id":    doc.ID,
			},
		}
		actionJSON, _ := json.Marshal(action)
		bulkBody.Write(actionJSON)
		bulkBody.WriteString("\n")
		
		// Document line
		docJSON, _ := json.Marshal(doc)
		bulkBody.Write(docJSON)
		bulkBody.WriteString("\n")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(bulkBody.String()))
	if err != nil {
		return fmt.Errorf("failed to create bulk request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-ndjson")
	req.SetBasicAuth(z.username, z.password)

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute bulk request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("zinc bulk request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// HealthCheck verifies Zinc is available
func (z *ZincRepository) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/index", z.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.SetBasicAuth(z.username, z.password)

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("zinc health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("zinc health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// ZincSearchRequest represents a Zinc search request
type ZincSearchRequest struct {
	Query map[string]interface{} `json:"query"`
	From  int                    `json:"from"`
	Size  int                    `json:"size"`
	Sort  []interface{}          `json:"sort,omitempty"`
}

// ZincSearchResponse represents a Zinc search response
type ZincSearchResponse struct {
	Took     int `json:"took"`
	TimedOut bool `json:"timed_out"`
	Hits     struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		MaxScore float64     `json:"max_score"`
		Hits     []ZincHit   `json:"hits"`
	} `json:"hits"`
}

// ZincHit represents a single search hit from Zinc
type ZincHit struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

// buildZincQuery converts domain SearchQuery to Zinc query format
func (z *ZincRepository) buildZincQuery(query domain.SearchQuery) ZincSearchRequest {
	zincQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []interface{}{
				map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  query.Query,
						"fields": []string{"title^2", "content", "tags"},
					},
				},
			},
		},
	}

	// Add filters
	var filters []interface{}
	
	if query.Source != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{"source": query.Source},
		})
	}
	
	if query.Type != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{"type": string(query.Type)},
		})
	}
	
	for key, value := range query.Filters {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{fmt.Sprintf("metadata.%s", key): value},
		})
	}

	if len(filters) > 0 {
		zincQuery["bool"].(map[string]interface{})["filter"] = filters
	}

	// Build sort
	var sort []interface{}
	if query.SortBy != "" {
		sortOrder := "desc"
		if query.SortOrder == "asc" {
			sortOrder = "asc"
		}
		sort = append(sort, map[string]interface{}{
			query.SortBy: map[string]string{"order": sortOrder},
		})
	}

	return ZincSearchRequest{
		Query: zincQuery,
		From:  query.Offset,
		Size:  query.Limit,
		Sort:  sort,
	}
}

// convertZincResponse converts Zinc response to domain SearchResults
func (z *ZincRepository) convertZincResponse(zincResp ZincSearchResponse, originalQuery domain.SearchQuery, took int) domain.SearchResults {
	results := make([]domain.SearchResult, len(zincResp.Hits.Hits))
	
	for i, hit := range zincResp.Hits.Hits {
		// Convert source back to SearchDocument
		doc := domain.SearchDocument{}
		if sourceBytes, err := json.Marshal(hit.Source); err == nil {
			json.Unmarshal(sourceBytes, &doc)
		}
		
		results[i] = domain.SearchResult{
			Document: doc,
			Score:    hit.Score,
			Snippet:  z.generateSnippet(doc, originalQuery.Query),
		}
	}

	total := zincResp.Hits.Total.Value
	pages := (total + originalQuery.Limit - 1) / originalQuery.Limit

	return domain.SearchResults{
		Results:  results,
		Total:    total,
		Query:    originalQuery.Query,
		Took:     took,
		MaxScore: zincResp.Hits.MaxScore,
		Pagination: domain.Pagination{
			Limit:  originalQuery.Limit,
			Offset: originalQuery.Offset,
			Total:  total,
			Pages:  pages,
		},
	}
}

// generateSnippet creates a highlighted excerpt from the document
func (z *ZincRepository) generateSnippet(doc domain.SearchDocument, query string) string {
	content := doc.Content
	if len(content) > 200 {
		content = content[:200] + "..."
	}
	return content
}