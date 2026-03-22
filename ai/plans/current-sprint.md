# Current Sprint

**Sprint goal:** ADR-012 — Refactor AFL domain entities into proper aggregates with domain purity

## Tasks

### ADR
- [x] Write ADR-012: Domain Purity and Aggregate Completeness
- [x] Add to decisions index

### Domain layer — aggregate modelling
- [x] Reshape `Match` as aggregate root: embed `Home`/`Away` as `ClubMatch` values (not IDs)
- [x] Reshape `ClubMatch`: embed `[]PlayerMatch` as child entities
- [x] Move `CalculateClubMatchScore` to method on `ClubMatch`
- [x] Add `Match.Winner()` domain method
- [x] Add domain unit tests for `ClubMatch.Score()`, `Match.Winner()`, edge cases (draw, no players)

### Domain layer — repository interfaces
- [x] Add aggregate-aware repository methods (e.g. `MatchRepository.FindByIDWithDetails`)
- [x] Keep existing per-entity methods for lightweight queries

### Infrastructure layer — aggregate repositories
- [x] Implement aggregate-loading repository methods (multi-query assembly)
- [x] Reuse existing sqlc queries (no new SQL needed)

### Application layer — use cases
- [x] Add `GetMatchWithDetails` query
- [x] Refactor `UpdatePlayerMatch` command to use aggregate for score recalculation

### Interface layer — resolvers
- [x] Update GraphQL resolvers to use `Home.ID`/`Away.ID` instead of removed `HomeClubMatchID`/`AwayClubMatchID`
- [x] Use `StoredScore` in convert layer

### Tests
- [x] Domain unit tests (pure logic, no mocks) — 16/16 pass
- [ ] Integration tests for aggregate repository methods (requires DB)
- [x] Verify existing GraphQL integration tests still compile and pass structurally
