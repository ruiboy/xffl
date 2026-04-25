package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	contractevents "xffl/contracts/events"
	"xffl/services/ffl/internal/application"
	pg "xffl/services/ffl/internal/infrastructure/postgres"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
	gql "xffl/services/ffl/internal/interface/graphql"
	pgevents "xffl/shared/events/pg"
)

// logLevelFromEnv returns slog.LevelDebug if LOG_LEVEL=debug, otherwise LevelInfo.
func logLevelFromEnv() slog.Level {
	if os.Getenv("LOG_LEVEL") == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevelFromEnv(),
	})).With(slog.String("service", "ffl")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
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

	queries := application.NewQueries(
		pg.NewClubRepository(q),
		pg.NewSeasonRepository(q),
		pg.NewRoundRepository(q),
		pg.NewMatchRepository(q),
		pg.NewClubSeasonRepository(q),
		pg.NewClubMatchRepository(q),
		pg.NewPlayerRepository(q),
		pg.NewPlayerMatchRepository(q),
		pg.NewPlayerSeasonRepository(q),
	)

	dispatcher := pgevents.New(pool, "xffl_events")

	db := pg.NewDB(pool)
	commands := application.NewCommands(db, dispatcher, application.EventRepos{
		Rounds:        pg.NewRoundRepository(q),
		PlayerSeasons: pg.NewPlayerSeasonRepository(q),
		PlayerMatches: pg.NewPlayerMatchRepository(q),
	})

	dispatcher.Subscribe(contractevents.PlayerMatchUpdated, commands.HandlePlayerMatchUpdated)
	go func() {
		if err := dispatcher.Listen(ctx); err != nil {
			slog.ErrorContext(ctx, "FFL event listener stopped", slog.Any("error", err))
		}
	}()

	resolver := &gql.Resolver{Queries: queries, Commands: commands}
	srv := handler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{Resolvers: resolver}))
	srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		ctx = pg.WithQueryCounter(ctx)
		rh := next(ctx)
		return func(ctx context.Context) *graphql.Response {
			resp := rh(ctx)
			op := graphql.GetOperationContext(ctx)
			slog.DebugContext(ctx, "db queries", slog.String("op", op.OperationName), slog.Int64("count", pg.QueryCount(ctx)))
			return resp
		}
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("/", playground.Handler("FFL", "/query"))
	mux.Handle("/query", srv)

	slog.InfoContext(ctx, "FFL service starting", slog.String("port", port))
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.ErrorContext(ctx, "FFL service failed", slog.Any("error", err))
		os.Exit(1)
	}
}
