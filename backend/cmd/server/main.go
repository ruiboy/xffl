package main

import (
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

	"xffl/internal/adapters/graphql"
	"xffl/internal/adapters/persistence"
	"xffl/internal/application"
	"xffl/internal/infrastructure"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize database connection
	database := infrastructure.NewDatabase()
	defer database.Close()

	// Initialize repositories
	clubRepo := persistence.NewClubRepository(database.DB)
	clubSeasonRepo := persistence.NewClubSeasonRepository(database.DB)
	playerRepo := persistence.NewPlayerRepository(database.DB)

	// Initialize use cases (application services)
	clubUseCase := application.NewClubService(clubRepo)
	clubSeasonUseCase := application.NewClubSeasonService(clubSeasonRepo)
	playerUseCase := application.NewPlayerService(playerRepo, clubRepo)

	// Initialize GraphQL resolver with dependency injection
	resolver := graphql.NewResolver(clubUseCase, playerUseCase, clubSeasonUseCase)

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
