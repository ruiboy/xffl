package zinc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client makes HTTP requests to a ZincSearch instance.
type Client struct {
	baseURL    string
	username   string
	password   string
	index      string
	httpClient *http.Client
}

// NewClient creates a Client targeting the given ZincSearch base URL.
func NewClient(baseURL, username, password string) *Client {
	return &Client{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		index:      "xffl",
		httpClient: &http.Client{},
	}
}

// EnsureIndex creates the index with keyword mappings for source and type so
// that term-query filtering works correctly. Safe to call on an existing index.
func (c *Client) EnsureIndex(ctx context.Context) error {
	body := map[string]any{
		"name":         c.index,
		"storage_type": "disk",
		"mappings": map[string]any{
			"properties": map[string]any{
				"source": map[string]any{"type": "keyword"},
				"type":   map[string]any{"type": "keyword"},
			},
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("zinc ensure index marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/index", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("zinc ensure index request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("zinc ensure index do: %w", err)
	}
	defer resp.Body.Close()

	// 200 = created, 400 may mean already exists — read and check.
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		// Zinc returns 400 with "index already exists" — treat as success.
		if resp.StatusCode == http.StatusBadRequest {
			return nil
		}
		return fmt.Errorf("zinc ensure index: status %d: %s", resp.StatusCode, b)
	}
	return nil
}

// indexDoc stores a document at PUT /api/{index}/_doc/{id}.
func (c *Client) indexDoc(ctx context.Context, docID string, body map[string]any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("zinc index marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/%s/_doc/%s", c.baseURL, c.index, docID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("zinc index request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("zinc index do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("zinc index: status %d: %s", resp.StatusCode, b)
	}
	return nil
}

// zincSearchResponse is the subset of the Zinc _search response we care about.
type zincSearchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string         `json:"_id"`
			Source map[string]any `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// search executes POST /api/{index}/_search with the given Zinc query DSL body.
func (c *Client) search(ctx context.Context, body map[string]any) (*zincSearchResponse, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("zinc search marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/%s/_search", c.baseURL, c.index)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("zinc search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zinc search do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("zinc search: status %d: %s", resp.StatusCode, b)
	}

	var result zincSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("zinc search decode: %w", err)
	}
	return &result, nil
}
