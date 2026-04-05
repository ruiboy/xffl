# Testing Conventions

Patterns and decisions for all Go tests in this repo. Read this before writing any tests.

## Stack

| Layer | Tools |
|-------|-------|
| Domain unit tests | `testing` + `testify` (assert/require) |
| Integration tests (DB) | `testing` + `testify` + `testcontainers-go` |

No other test libraries. Do not add mocks for the database — integration tests hit a real Postgres
instance via testcontainers.

---

## Domain unit tests

Pure functions in `internal/domain/` get table-driven unit tests. No infrastructure, no containers.

**Pattern:**

```go
func TestClubMatch_Score(t *testing.T) {
    tests := []struct {
        name string
        cm   ClubMatch
        want int
    }{
        {"empty match scores zero", ClubMatch{}, 0},
        {"rushed behinds contribute to score without players", ClubMatch{RushedBehinds: 3}, 3},
        {"single player goals and behinds are summed correctly", ClubMatch{
            PlayerMatches: []PlayerMatch{{Goals: 2, Behinds: 1}},
        }, 13},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.Equal(t, tt.want, tt.cm.Score())
        })
    }
}
```

**Rules:**
- Test function name: `Test<Type>_<Method>` — identifies what is under test
- Table case names: short sentence stating the expectation ("empty match scores zero")
- Use `assert.Equal(t, want, got)` — want first, got second
- No `t.Parallel()` — domain tests are microsecond-fast; the overhead and added complexity is not worth it

---

## Integration tests (DB-backed)

Tests in `internal/interface/graphql/` (and any future DB-backed packages) use a real Postgres
container via testcontainers-go.

### Container lifecycle — `TestMain`

Each package that needs a DB gets a `testmain_test.go`:

```go
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
```

- Container starts **once per package** (`TestMain` is per-package in Go)
- `testPool` is a package-level var shared by all test functions in the package
- Do **not** use `WithReuse(true)` — fresh container per `go test` run keeps tests hermetic
- If a second package needs integration tests, give it its own `testmain_test.go` calling `testutil.StartPostgres`

### testutil package

`services/afl/internal/testutil/postgres.go` provides `StartPostgres(ctx)`:
- Spins up `postgres:16-alpine`
- Walks up the directory tree to find the repo root (looks for `justfile`)
- Applies `dev/postgres/init/01_afl_schema.sql` via `WithInitScripts`
- Returns `(*pgxpool.Pool, func(), error)` — caller calls cleanup after `m.Run()`

### Per-test data isolation

Each test calls a `seedTestData(t, pool)` helper that:
1. **Truncates all AFL tables** (in FK order) at the start — clears state from any previous test
2. Inserts the test graph
3. Registers `t.Cleanup` to truncate again after the test

This means tests are isolated even though they share the same container.

### Assertion style

```go
// Fatal (stops the test immediately) — use require:
require.Empty(t, result.Errors)
require.NoError(t, json.Unmarshal(result.Data, &data))
require.Len(t, data.Rounds, 1)          // when you're about to index into it
require.NotNil(t, match.HomeClubMatch)  // before dereferencing a pointer

// Non-fatal (continues after failure) — use assert:
assert.Equal(t, "Test Cats", data.AflClubs[0].Name)
assert.Len(t, data.AflClubs, 2)
assert.Empty(t, data.Results)
```

Use `require` for setup/structural checks where continuing would panic. Use `assert` for value
checks where all failures are useful.

### t.Run naming

Group related assertions under `t.Run` with a sentence that states the expectation:

```go
t.Run("clubs ordered alphabetically", func(t *testing.T) {
    assert.Equal(t, "Mountain Goats", data.AflClubs[0].Name)
    assert.Equal(t, "Sky Pilots", data.AflClubs[1].Name)
})
t.Run("ladder ordered by premiership points descending", func(t *testing.T) {
    assert.Equal(t, "Sky Pilots", data.AflSeason.Ladder[0].Club.Name)
    assert.Equal(t, 16, data.AflSeason.Ladder[0].PremiershipPoints)
})
```

Top-level function names identify the feature/query/mutation (`TestAflClubs`, `TestUpdateAFLPlayerMatch_Update`).
`t.Run` names state what specifically is expected ("clubs ordered alphabetically").

### connectDB helper

Each integration test file exposes a thin `connectDB` helper that returns the shared pool:

```go
func connectDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    return testPool
}
```

No skip logic, no connection setup — `TestMain` owns the pool lifecycle.

---

## Running tests

```bash
# Domain unit tests (no Docker needed)
cd services/afl && go test ./internal/domain/...

# Integration tests (requires Docker)
cd services/afl && go test ./internal/interface/graphql/... -v

# All tests
cd services/afl && go test ./...
```
