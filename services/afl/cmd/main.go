package main

import (
	"context"
	"fmt"
	"log/slog"
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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevelFromEnv(),
	})))

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
		slog.ErrorContext(ctx, "unable to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	q := sqlcgen.New(pool)

	clk := clockFromEnv(ctx)

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
			slog.ErrorContext(ctx, "AFL event listener stopped", slog.Any("error", err))
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

	slog.InfoContext(ctx, "AFL service starting", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.ErrorContext(ctx, "AFL service failed", slog.Any("error", err))
		os.Exit(1)
	}
}

// logLevelFromEnv returns slog.LevelDebug if LOG_LEVEL=debug, otherwise LevelInfo.
func logLevelFromEnv() slog.Level {
	if os.Getenv("LOG_LEVEL") == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

// clockFromEnv returns a FixedClock if CLOCK_OVERRIDE is set (for e2e tests),
// otherwise a RealClock.
func clockFromEnv(ctx context.Context) clock.Clock {
	if override := os.Getenv("CLOCK_OVERRIDE"); override != "" {
		t, err := time.Parse(time.RFC3339, override)
		if err != nil {
			slog.ErrorContext(ctx, "invalid CLOCK_OVERRIDE", slog.String("value", override), slog.Any("error", err))
			os.Exit(1)
		}
		slog.InfoContext(ctx, "AFL clock overridden", slog.String("time", t.Format(time.RFC3339)))
		return clock.FixedClock{T: t}
	}
	return clock.RealClock{}
}
