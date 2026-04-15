package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/jackc/pgx/v5/pgxpool"

	contractevents "xffl/contracts/events"
	"xffl/services/search/internal/application"
	"xffl/services/search/internal/infrastructure/typesense"
	gql "xffl/services/search/internal/interface/graphql"
	pgevents "xffl/shared/events/pg"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevelFromEnv(),
	})))

	ctx := context.Background()

	port := envOr("PORT", "8082")
	dbURL := envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/xffl?sslmode=disable")
	tsHost := envOr("TYPESENSE_HOST", "localhost")
	tsPort := envOr("TYPESENSE_PORT", "8108")
	tsAPIKey := envOr("TYPESENSE_API_KEY", "xyz")

	// Postgres — used only for event listening, no schema.
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.ErrorContext(ctx, "unable to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	// Typesense
	tsClient := typesense.NewClient(fmt.Sprintf("http://%s:%s", tsHost, tsPort), tsAPIKey)
	repo := typesense.NewRepository(tsClient)
	if err := repo.EnsureCollection(ctx); err != nil {
		slog.ErrorContext(ctx, "ensure typesense collection", slog.Any("error", err))
		os.Exit(1)
	}

	// Application use cases
	indexUC := application.NewIndexDocument(repo)
	handlers := application.NewHandlers(indexUC)

	// Event subscriptions
	dispatcher := pgevents.New(pool, "xffl_events")
	dispatcher.Subscribe(contractevents.PlayerMatchUpdated, handlers.HandlePlayerMatchUpdated)
	dispatcher.Subscribe(contractevents.FantasyScoreCalculated, handlers.HandleFantasyScoreCalculated)
	go func() {
		if err := dispatcher.Listen(ctx); err != nil {
			slog.ErrorContext(ctx, "search event listener stopped", slog.Any("error", err))
		}
	}()

	// GraphQL
	resolver := &gql.Resolver{Repo: repo}
	srv := handler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("/", playground.Handler("Search", "/query"))
	mux.Handle("/query", srv)

	slog.InfoContext(ctx, "Search service starting", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.ErrorContext(ctx, "Search service failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func logLevelFromEnv() slog.Level {
	if os.Getenv("LOG_LEVEL") == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
