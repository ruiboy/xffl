package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"xffl/services/afl/internal/application"
	pg "xffl/services/afl/internal/infrastructure/postgres"
	"xffl/services/afl/internal/infrastructure/postgres/sqlcgen"
	gql "xffl/services/afl/internal/interface/graphql"
	"xffl/shared/clock"
	pgevents "xffl/shared/events/pg"
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

	q := sqlcgen.New(pool)

	clk := clockFromEnv()

	queries := application.NewQueries(
		clk,
		pg.NewClubRepository(q),
		pg.NewSeasonRepository(q),
		pg.NewRoundRepository(q, pool),
		pg.NewMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewPlayerRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)

	dispatcher := pgevents.New(pool, "xffl_events")
	go func() {
		if err := dispatcher.Listen(ctx); err != nil {
			log.Printf("AFL: event listener stopped: %v", err)
		}
	}()

	db := pg.NewDB(pool)
	commands := application.NewCommands(db, dispatcher)

	resolver := &gql.Resolver{Queries: queries, Commands: commands}
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

// clockFromEnv returns a FixedClock if CLOCK_OVERRIDE is set (for e2e tests),
// otherwise a RealClock.
func clockFromEnv() clock.Clock {
	if override := os.Getenv("CLOCK_OVERRIDE"); override != "" {
		t, err := time.Parse(time.RFC3339, override)
		if err != nil {
			log.Fatalf("invalid CLOCK_OVERRIDE %q: %v", override, err)
		}
		log.Printf("AFL: clock overridden to %s", t.Format(time.RFC3339))
		return clock.FixedClock{T: t}
	}
	return clock.RealClock{}
}
