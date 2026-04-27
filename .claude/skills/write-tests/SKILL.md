---
name: write-tests
description: Write Go tests following project conventions (testify, testcontainers, expressive t.Run names)
---

Read `ai/architecture/testing.md` first. Then write tests for the code the user specifies.

Steps:
1. Identify the layer under test:
   - `internal/domain/` → table-driven unit test (no DB, no containers)
   - `internal/infrastructure/` (parsers, resolvers, adapters with no DB) → plain unit test; same testify rules as domain, table-driven where inputs vary
   - `internal/interface/graphql/` or any DB-backed layer → integration test with testcontainers

2. For **unit tests** (domain or infrastructure):
   - Use `Test<Type>_<Method>` as the function name; for file-driven tests `Test<Format>` is fine
   - Use `require` for fatal preconditions (file read, nil guards before indexing)
   - Use `assert.Equal(t, want, got)` for value checks — want first, got second
   - Table-driven where inputs vary; single `TestX` function with named sub-checks where one input exercises multiple assertions
   - Do not add `t.Parallel()`

3. For **integration tests**:
   - Check whether a `testmain_test.go` already exists in the package. If not, create one using the `TestMain` pattern from `ai/architecture/testing.md`
   - Use `connectDB(t)` + `seedTestData(t, pool)` + `setupTestServer(t, pool)` as the test preamble
   - Use `require` for fatal checks (no errors, JSON decode, nil guards before indexing)
   - Use `assert` inside `t.Run` blocks for value expectations
   - Each `t.Run` name is a sentence stating what is expected

4. Follow existing examples exactly:
   - Domain: `services/afl/internal/domain/club_match_test.go`
   - Infrastructure unit: `services/ffl/internal/infrastructure/forum/parser_test.go`
   - Integration: `services/afl/internal/interface/graphql/integration_test.go`

Do not invent new patterns. If the conventions doc does not cover the situation, flag it and ask before proceeding.