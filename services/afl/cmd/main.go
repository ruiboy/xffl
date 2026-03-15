package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"xffl/services/afl/internal/application"
	pg "xffl/services/afl/internal/infrastructure/postgres"
	gql "xffl/services/afl/internal/interface/graphql"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/xffl?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pg.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	queries := application.NewQueries(
		pg.NewClubRepository(pool),
		pg.NewSeasonRepository(pool),
		pg.NewRoundRepository(pool),
		pg.NewMatchRepository(pool),
		pg.NewClubSeasonRepository(pool),
		pg.NewClubMatchRepository(pool),
		pg.NewPlayerRepository(pool),
		pg.NewPlayerMatchRepository(pool),
		pg.NewPlayerSeasonRepository(pool),
	)

	resolver := &gql.Resolver{Queries: queries}
	srv := handler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("/", playground.Handler("AFL", "/query"))
	mux.Handle("/query", srv)

	log.Printf("AFL service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
