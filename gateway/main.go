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
	Data   interface{}    `json:"data,omitempty"`
	Errors []interface{}  `json:"errors,omitempty"`
}

type ServiceConfig struct {
	Services    map[string]ServiceDetails `json:"services"`
	BuildTime   string                    `json:"buildTime"`
	Version     string                    `json:"version"`
	TypeClashes []interface{}             `json:"typeClashes"`
}

type ServiceDetails struct {
	URL       string   `json:"url"`
	Queries   []string `json:"queries"`
	Mutations []string `json:"mutations"`
	Types     []string `json:"types"`
}

type Gateway struct {
	config    ServiceConfig
	startTime time.Time
}

func NewGateway() (*Gateway, error) {
	// Load service configuration
	configBytes, err := os.ReadFile("generated/service-config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read service config: %w", err)
	}

	var config ServiceConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to parse service config: %w", err)
	}

	return &Gateway{
		config:    config,
		startTime: time.Now(),
	}, nil
}

func (g *Gateway) routeRequest(query string) string {
	// Use config-based routing
	queryLower := strings.ToLower(query)
	
	// Gateway metadata - handle locally
	if strings.Contains(queryLower, "_gateway") {
		return ""
	}
	
	// Check each service's queries and mutations
	for _, details := range g.config.Services {
		// Check queries
		for _, queryName := range details.Queries {
			if strings.Contains(queryLower, strings.ToLower(queryName)) {
				// Skip introspection fields
				if strings.HasPrefix(queryName, "__") {
					continue
				}
				return details.URL
			}
		}
		// Check mutations  
		for _, mutationName := range details.Mutations {
			if strings.Contains(queryLower, strings.ToLower(mutationName)) {
				return details.URL
			}
		}
	}
	
	// Default to first service if no match found
	for _, details := range g.config.Services {
		return details.URL
	}
	
	return ""
}

func (g *Gateway) handleGatewayMetadata() GraphQLResponse {
	var services []string
	for serviceName := range g.config.Services {
		services = append(services, serviceName)
	}

	return GraphQLResponse{
		Data: map[string]interface{}{
			"_gateway": map[string]interface{}{
				"version":   g.config.Version,
				"services":  services,
				"lastBuild": g.config.BuildTime,
				"uptime":    time.Since(g.startTime).Seconds(),
			},
		},
	}
}

func (g *Gateway) proxyToService(serviceURL string, req GraphQLRequest) (*GraphQLResponse, error) {
	// Marshal request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBytes, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &gqlResp, nil
}

func (g *Gateway) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
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

	// Parse request
	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Check if this is a gateway metadata query
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

	// Return response
	json.NewEncoder(w).Encode(gqlResp)
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	gateway, err := NewGateway()
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/query", gateway.handleGraphQL)
	mux.HandleFunc("/health", gateway.handleHealth)

	port := getEnvOrDefault("PORT", "8090")
	
	fmt.Printf("🚀 XFFL Gateway starting on port %s\n", port)
	fmt.Printf("📊 GraphQL endpoint: http://localhost:%s/query\n", port)
	fmt.Printf("🔍 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔗 Loaded %d services from config\n", len(gateway.config.Services))
	for serviceName, details := range gateway.config.Services {
		fmt.Printf("   - %s: %s (%d queries, %d mutations)\n", 
			serviceName, details.URL, len(details.Queries), len(details.Mutations))
	}
	
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}