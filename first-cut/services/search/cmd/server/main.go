package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pkgEvents "xffl/pkg/events/postgres"
	"xffl/services/search/internal/adapters/events"
	httpHandlers "xffl/services/search/internal/adapters/http"
	"xffl/services/search/internal/adapters/zinc"
	"xffl/services/search/internal/services"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "[SEARCH] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting XFFL Search Service...")

	// Load configuration from environment
	config := loadConfig()
	logger.Printf("Loaded configuration: Port=%s, ZincURL=%s", config.Port, config.ZincURL)

	// Database connection is handled by the PostgreSQL event dispatcher
	logger.Println("Database connection managed by event dispatcher")

	// Initialize event dispatcher
	eventDispatcher, err := pkgEvents.NewPostgresDispatcher(config.EventDBURL, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize event dispatcher: %v", err)
	}

	// Start event dispatcher
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eventDispatcher.Start(ctx); err != nil {
		logger.Fatalf("Failed to start event dispatcher: %v", err)
	}
	logger.Println("Event dispatcher started")

	// Initialize Zinc search repository
	zincRepo := zinc.NewZincRepository(zinc.ZincConfig{
		BaseURL:   config.ZincURL,
		Username:  config.ZincUsername,
		Password:  config.ZincPassword,
		IndexName: config.ZincIndexName,
		Timeout:   30 * time.Second,
	})

	// Test Zinc connection
	if err := zincRepo.HealthCheck(ctx); err != nil {
		logger.Printf("Warning: Zinc health check failed: %v", err)
		logger.Println("Search service will start but search operations may fail")
	} else {
		logger.Println("Connected to Zinc search engine")
	}

	// Initialize services
	searchService := services.NewSearchService(zincRepo)
	indexingService := services.NewIndexingService(searchService, eventDispatcher, logger)

	// Initialize and register event handlers
	playerMatchHandler := events.NewPlayerMatchHandler(indexingService, logger)
	fantasyScoreHandler := events.NewFantasyScoreHandler(indexingService, logger)

	// Subscribe to relevant events
	if err := eventDispatcher.Subscribe("AFL.PlayerMatchUpdated", playerMatchHandler); err != nil {
		logger.Printf("Failed to subscribe to AFL.PlayerMatchUpdated events: %v", err)
	}

	if err := eventDispatcher.Subscribe("FFL.FantasyScoreCalculated", fantasyScoreHandler); err != nil {
		logger.Printf("Failed to subscribe to FFL.FantasyScoreCalculated events: %v", err)
	}

	logger.Println("Event handlers registered")

	// Initialize HTTP handlers
	searchHandler := httpHandlers.NewSearchHandler(searchService)
	indexHandler := httpHandlers.NewIndexHandler(searchService)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Public search endpoints
	mux.HandleFunc("/search", searchHandler.HandleSearch)
	mux.HandleFunc("/health", searchHandler.HandleHealthCheck)

	// Admin endpoints for manual indexing
	mux.HandleFunc("/admin/index", indexHandler.HandleIndexDocument)
	mux.HandleFunc("/admin/index/", indexHandler.HandleDeleteDocument)

	// Setup server
	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("🔍 XFFL Search Service starting on port %s", config.Port)
		logger.Printf("📊 Search endpoint: http://localhost:%s/search", config.Port)
		logger.Printf("🔍 Health check: http://localhost:%s/health", config.Port)
		logger.Printf("⚙️  Admin indexing: http://localhost:%s/admin/index", config.Port)
		logger.Printf("🔗 Zinc URL: %s", config.ZincURL)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down search service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Error during server shutdown: %v", err)
	}

	// Stop event dispatcher
	if err := eventDispatcher.Stop(); err != nil {
		logger.Printf("Error stopping event dispatcher: %v", err)
	}

	logger.Println("Search service stopped")
}

// Config holds the application configuration
type Config struct {
	Port          string
	DBHost        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBPort        string
	EventDBURL    string
	ZincURL       string
	ZincUsername  string
	ZincPassword  string
	ZincIndexName string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		Port:          getEnvOrDefault("PORT", "8082"),
		DBHost:        getEnvOrDefault("DB_HOST", "localhost"),
		DBUser:        getEnvOrDefault("DB_USER", "postgres"),
		DBPassword:    getEnvOrDefault("DB_PASSWORD", ""),
		DBName:        getEnvOrDefault("DB_NAME", "xffl"),
		DBPort:        getEnvOrDefault("DB_PORT", "5432"),
		EventDBURL:    getEnvOrDefault("EVENT_DB_URL", "user=postgres dbname=xffl sslmode=disable"),
		ZincURL:       getEnvOrDefault("ZINC_URL", "http://localhost:4080"),
		ZincUsername:  getEnvOrDefault("ZINC_USERNAME", "admin"),
		ZincPassword:  getEnvOrDefault("ZINC_PASSWORD", "admin"),
		ZincIndexName: getEnvOrDefault("ZINC_INDEX_NAME", "xffl"),
	}
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
