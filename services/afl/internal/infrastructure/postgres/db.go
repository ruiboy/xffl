package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/application"
	"xffl/services/afl/internal/infrastructure/postgres/sqlcgen"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	config.ConnConfig.Tracer = QueryCountTracer{}
	return pgxpool.NewWithConfig(ctx, config)
}

// DB provides transactional access to repositories.
type DB struct {
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

func (db *DB) WithTx(ctx context.Context, fn func(repos application.WriteRepos) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txQ := sqlcgen.New(tx)
	repos := application.WriteRepos{
		Players:       NewPlayerRepository(txQ),
		PlayerSeasons: NewPlayerSeasonRepository(txQ),
		PlayerMatches: NewPlayerMatchRepository(txQ),
		ClubMatches:   NewClubMatchRepository(txQ),
	}

	if err := fn(repos); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
