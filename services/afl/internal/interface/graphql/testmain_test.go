package graphql_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"xffl/services/afl/internal/testutil"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	pool, cleanup, err := testutil.StartPostgres(ctx)
	if err != nil {
		panic(err)
	}
	testPool = pool
	code := m.Run()
	cleanup()
	os.Exit(code)
}
