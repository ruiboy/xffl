package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"
	"github.com/vektah/gqlparser/v2/ast"

	"xffl/pkg/database"
	"xffl/pkg/events/postgres"
	"xffl/services/afl/internal/adapters/graphql"
	dbadapter "xffl/services/afl/internal/adapters/db"
	"xffl/services/afl/internal/services"
)

const defaultPort = "8081"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize database connection
	db := database.NewDatabase()
	defer db.Close()

	// Initialize PostgreSQL event dispatcher (separate from domain persistence)
	eventLogger := log.New(os.Stdout, "[AFL-EVENTS] ", log.LstdFlags)
	eventConnStr := getEnvOrDefault("EVENT_DB_URL", "user=postgres dbname=xffl sslmode=disable")
	eventDispatcher, err := postgres.NewPostgresDispatcher(eventConnStr, eventLogger)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL event dispatcher: %v", err)
	}

	// Start event dispatcher
	ctx := context.Background()
	if err := eventDispatcher.Start(ctx); err != nil {
		log.Fatalf("Failed to start event dispatcher: %v", err)
	}
	defer func() {
		if err := eventDispatcher.Stop(); err != nil {
			log.Printf("Error stopping event dispatcher: %v", err)
		}
	}()

	// Initialize repositories
	clubRepo := dbadapter.NewClubRepository(db.DB)
	playerMatchRepo := dbadapter.NewPlayerMatchRepository(db.DB)

	// Initialize services
	clubService := services.NewClubService(clubRepo)
	playerMatchService := services.NewPlayerMatchService(playerMatchRepo, eventDispatcher)

	// Initialize GraphQL resolver with dependency injection
	resolver := graphql.NewResolver(clubService, playerMatchService)

	// Initialize GraphQL server
	srv := handler.New(graphql.NewExecutableSchema(graphql.Config{Resolvers: resolver}))

	// Add transports
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	// Configure caching
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Add extensions
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Create a new mux router
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", srv)

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap the mux with CORS middleware
	handler := c.Handler(mux)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
