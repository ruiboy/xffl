Read `ai/architecture/testing.md` first. Then write tests for the code the user specifies.

Steps:
1. Identify the layer under test:
   - `internal/domain/` → table-driven unit test (no DB, no containers)
   - `internal/interface/graphql/` or any DB-backed layer → integration test with testcontainers

2. For **domain unit tests**:
   - Use `Test<Type>_<Method>` as the function name
   - Build a `tests []struct{ name string; ... }` table
   - Each case name is a short sentence stating the expectation ("empty match scores zero")
   - Use `assert.Equal(t, want, got)` inside `t.Run`
   - Do not add `t.Parallel()`

3. For **integration tests**:
   - Check whether a `testmain_test.go` already exists in the package. If not, create one using the `TestMain` pattern from `ai/architecture/testing.md`
   - Use `connectDB(t)` + `seedTestData(t, pool)` + `setupTestServer(t, pool)` as the test preamble
   - Use `require` for fatal checks (no errors, JSON decode, nil guards before indexing)
   - Use `assert` inside `t.Run` blocks for value expectations
   - Each `t.Run` name is a sentence stating what is expected

4. Follow existing examples exactly:
   - Domain: `services/afl/internal/domain/club_match_test.go`
   - Integration: `services/afl/internal/interface/graphql/integration_test.go`

Do not invent new patterns. If the conventions doc does not cover the situation, flag it and ask before proceeding.