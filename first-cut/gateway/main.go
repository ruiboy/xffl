package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
}

type Gateway struct {
	aflServiceURL    string
	fflServiceURL    string
	searchServiceURL string
	startTime        time.Time
}

func NewGateway() *Gateway {
	return &Gateway{
		aflServiceURL:    getEnvOrDefault("AFL_SERVICE_URL", "http://localhost:8080/query"),
		fflServiceURL:    getEnvOrDefault("FFL_SERVICE_URL", "http://localhost:8081/query"),
		searchServiceURL: getEnvOrDefault("SEARCH_SERVICE_URL", "http://localhost:8082"),
		startTime:        time.Now(),
	}
}

func (g *Gateway) routeRequest(query string) string {
	queryLower := strings.ToLower(query)

	// Route based on service prefix in query/mutation names
	if strings.Contains(queryLower, "afl") {
		return g.aflServiceURL
	}
	
	if strings.Contains(queryLower, "ffl") {
		return g.fflServiceURL
	}

	// Gateway metadata - handle locally
	if strings.Contains(queryLower, "_gateway") {
		return ""
	}

	// Default to FFL service
	return g.fflServiceURL
}

func (g *Gateway) handleGatewayMetadata() GraphQLResponse {
	return GraphQLResponse{
		Data: map[string]interface{}{
			"_gateway": map[string]interface{}{
				"version":   "1.0.0",
				"services":  []string{"afl", "ffl"},
				"lastBuild": g.startTime.Format(time.RFC3339),
				"uptime":    time.Since(g.startTime).Seconds(),
			},
		},
	}
}

func (g *Gateway) proxyToService(serviceURL string, req GraphQLRequest) (*GraphQLResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBytes, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &gqlResp, nil
}

func (g *Gateway) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Handle gateway metadata locally
	if strings.Contains(strings.ToLower(req.Query), "_gateway") {
		resp := g.handleGatewayMetadata()
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Route to appropriate service
	serviceURL := g.routeRequest(req.Query)
	if serviceURL == "" {
		http.Error(w, "Could not route request", http.StatusBadRequest)
		return
	}

	// Proxy to service
	gqlResp, err := g.proxyToService(serviceURL, req)
	if err != nil {
		log.Printf("Proxy error: %v", err)
		errorResp := GraphQLResponse{
			Errors: []interface{}{map[string]string{"message": err.Error()}},
		}
		json.NewEncoder(w).Encode(errorResp)
		return
	}

	json.NewEncoder(w).Encode(gqlResp)
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (g *Gateway) handleSearch(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Proxy search request to search service
	searchURL := fmt.Sprintf("%s/search?%s", g.searchServiceURL, r.URL.RawQuery)
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil {
		log.Printf("Search proxy error: %v", err)
		http.Error(w, "Search service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy status code
	w.WriteHeader(resp.StatusCode)
	
	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response body
	io.Copy(w, resp.Body)
}

func main() {
	gateway := NewGateway()

	mux := http.NewServeMux()
	mux.HandleFunc("/query", gateway.handleGraphQL)
	mux.HandleFunc("/search", gateway.handleSearch)
	mux.HandleFunc("/health", gateway.handleHealth)

	port := getEnvOrDefault("PORT", "8090")

	fmt.Printf("🚀 XFFL Minimal Gateway starting on port %s\n", port)
	fmt.Printf("📊 GraphQL endpoint: http://localhost:%s/query\n", port)
	fmt.Printf("🔍 Search endpoint: http://localhost:%s/search\n", port)
	fmt.Printf("🔍 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔗 AFL Service: %s\n", gateway.aflServiceURL)
	fmt.Printf("🔗 FFL Service: %s (default)\n", gateway.fflServiceURL)
	fmt.Printf("🔗 Search Service: %s\n", gateway.searchServiceURL)
	fmt.Printf("🎯 Routing: 'afl' → AFL, 'ffl' → FFL, default → FFL\n")

	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}