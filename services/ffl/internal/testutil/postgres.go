package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartPostgres starts a postgres:16-alpine container with the AFL and FFL schemas applied.
// FFL players reference afl.player, so both schemas are required.
// Returns the pool, a cleanup func, and any error.
func StartPostgres(ctx context.Context) (*pgxpool.Pool, func(), error) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return nil, nil, fmt.Errorf("find repo root: %w", err)
	}
	aflSchema := filepath.Join(repoRoot, "dev", "postgres", "init", "01_afl_schema.sql")
	fflSchema := filepath.Join(repoRoot, "dev", "postgres", "init", "02_ffl_schema.sql")
	dataops := filepath.Join(repoRoot, "dev", "postgres", "init", "03_dataops.sql")

	ctr, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("xffl"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.WithInitScripts(aflSchema, fflSchema, dataops),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("start postgres container: %w", err)
	}

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		ctr.Terminate(ctx) //nolint:errcheck
		return nil, nil, fmt.Errorf("get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		ctr.Terminate(ctx) //nolint:errcheck
		return nil, nil, fmt.Errorf("create pool: %w", err)
	}

	cleanup := func() {
		pool.Close()
		ctr.Terminate(ctx) //nolint:errcheck
	}
	return pool, cleanup, nil
}

// findRepoRoot walks up from cwd until it finds a justfile (the repo root marker).
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "justfile")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("justfile not found walking up from %s", dir)
		}
		dir = parent
	}
}
