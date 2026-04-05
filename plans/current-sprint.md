# Current Sprint

**Sprint goal:** Phase 10 - Test stabilisation

Refactor tests to be accurate, extensible, and grounded in minimal seed data

## Completed

- [x] Decided against `t.Parallel()` for domain tests (microsecond-fast, no benefit)
- [x] Migrated AFL integration tests from dev Postgres to testcontainers (`services/afl/internal/testutil/postgres.go`)
- [x] Added `TestMain` + shared container pool pattern to AFL graphql test package
- [x] Rewrote AFL `integration_test.go` with testify (`require`/`assert`) and sentence-style `t.Run` names
- [x] Added `ai/architecture/testing.md` — full conventions doc (stack, patterns, naming, examples)
- [x] Added `/write-tests` skill at `.claude/skills/write-tests/SKILL.md`
- [x] Updated `ai/architecture/cookbook.md` testing section to point to testing.md
- [x] Consolidated FFL domain.md position/slot tables into one
- [x] Migrated FFL domain tests to testify with expressive names (`club_match_test.go`, `club_season_test.go`, `player_match_test.go`)
- [x] Migrated FFL integration tests to testcontainers (`services/ffl/internal/testutil/postgres.go`, `TestMain`)
- [x] Rewrote FFL `integration_test.go` with testify and sentence-style `t.Run` names
- [x] Migrated missing `commands_test.go` scenarios to integration tests (`addFFLSquadPlayer`, empty team, star position)
- [x] Deleted `services/ffl/internal/application/commands_test.go` — all coverage now end-to-end via testcontainers