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
	"xffl/pkg/events/memory"
	"xffl/services/afl/internal/adapters/graphql"
	"xffl/services/afl/internal/adapters/persistence"
	"xffl/services/afl/internal/application"
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

	// Initialize event dispatcher
	eventLogger := log.New(os.Stdout, "[AFL-EVENTS] ", log.LstdFlags)
	eventDispatcher := memory.NewInMemoryDispatcher(eventLogger)
	
	// Start event dispatcher
	ctx := context.Background()
	if err := eventDispatcher.Start(ctx); err != nil {
		log.Fatalf("Failed to start event dispatcher: %v", err)
	}
	defer eventDispatcher.Stop()

	// Initialize repositories
	clubRepo := persistence.NewClubRepository(db.DB)
	playerMatchRepo := persistence.NewPlayerMatchRepository(db.DB)

	// Initialize use cases (application services)
	clubUseCase := application.NewClubService(clubRepo)
	playerMatchUseCase := application.NewPlayerMatchService(playerMatchRepo, eventDispatcher)

	// Initialize GraphQL resolver with dependency injection
	resolver := graphql.NewResolver(clubUseCase, playerMatchUseCase)

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
