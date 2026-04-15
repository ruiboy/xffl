package typesense

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client makes HTTP requests to a Typesense instance.
type Client struct {
	apiURL     string // e.g. "http://localhost:8108"
	apiKey     string
	collection string
	httpClient *http.Client
}

// NewClient creates a Client targeting the given Typesense API URL.
func NewClient(apiURL, apiKey string) *Client {
	return &Client{
		apiURL:     strings.TrimRight(apiURL, "/"),
		apiKey:     apiKey,
		collection: "documents",
		httpClient: &http.Client{},
	}
}

// EnsureCollection creates the collection with the expected schema.
// Returns nil if the collection already exists.
func (c *Client) EnsureCollection(ctx context.Context) error {
	schema := map[string]any{
		"name": c.collection,
		"fields": []map[string]any{
			{"name": "source", "type": "string", "facet": true},
			{"name": "type", "type": "string", "facet": true},
			{"name": ".*", "type": "auto"},
		},
	}
	data, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("typesense ensure collection marshal: %w", err)
	}

	u := fmt.Sprintf("%s/collections", c.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("typesense ensure collection request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("typesense ensure collection do: %w", err)
	}
	defer resp.Body.Close()

	// 409 = collection already exists — treat as success.
	if resp.StatusCode == http.StatusConflict {
		return nil
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("typesense ensure collection: status %d: %s", resp.StatusCode, b)
	}
	return nil
}

// upsertDoc inserts or updates a document in the collection.
func (c *Client) upsertDoc(ctx context.Context, doc map[string]any) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("typesense upsert marshal: %w", err)
	}

	u := fmt.Sprintf("%s/collections/%s/documents?action=upsert", c.apiURL, c.collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("typesense upsert request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("typesense upsert do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("typesense upsert: status %d: %s", resp.StatusCode, b)
	}
	return nil
}

// searchResponse is the subset of the Typesense search response we care about.
type searchResponse struct {
	Found int `json:"found"`
	Hits  []struct {
		Document map[string]any `json:"document"`
	} `json:"hits"`
}

// search executes a search against the collection.
func (c *Client) search(ctx context.Context, q, queryBy, filterBy string) (*searchResponse, error) {
	params := url.Values{}
	params.Set("q", q)
	params.Set("query_by", queryBy)
	if filterBy != "" {
		params.Set("filter_by", filterBy)
	}
	params.Set("per_page", "250")

	u := fmt.Sprintf("%s/collections/%s/documents/search?%s", c.apiURL, c.collection, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("typesense search request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("typesense search do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("typesense search: status %d: %s", resp.StatusCode, b)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("typesense search decode: %w", err)
	}
	return &result, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TYPESENSE-API-KEY", c.apiKey)
}
