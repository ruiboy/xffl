# Current Sprint

**Sprint goal:** ADR-012 — Refactor AFL domain entities into proper aggregates with domain purity

## Tasks

### ADR
- [x] Write ADR-012: Domain Purity and Aggregate Completeness
- [x] Add to decisions index

### Domain layer — aggregate modelling
- [ ] Reshape `Match` as aggregate root: embed `Home`/`Away` as `ClubMatch` values (not IDs)
- [ ] Reshape `ClubMatch`: embed `[]PlayerMatch` as child entities
- [ ] Move `CalculateClubMatchScore` to method on `ClubMatch`
- [ ] Add `Match.Winner()` domain method
- [ ] Add domain unit tests for `ClubMatch.Score()`, `Match.Winner()`, edge cases (draw, no players)

### Domain layer — repository interfaces
- [ ] Add aggregate-aware repository methods (e.g. `MatchRepository.FindByIDWithDetails`)
- [ ] Keep existing per-entity methods for lightweight queries

### Infrastructure layer — aggregate repositories
- [ ] Implement aggregate-loading repository methods (JOIN queries or multi-query assembly)
- [ ] Add sqlc queries to support aggregate loading
- [ ] Regenerate sqlc code

### Application layer — use cases
- [ ] Refactor `Queries` to return aggregates where appropriate
- [ ] Refactor `UpdatePlayerMatch` command to use aggregate for score recalculation

### Interface layer — resolvers
- [ ] Update GraphQL resolvers to use aggregates from application layer
- [ ] Simplify field resolvers that currently do manual assembly

### Tests
- [ ] Domain unit tests (pure logic, no mocks)
- [ ] Integration tests for aggregate repository methods
- [ ] Verify existing GraphQL integration tests still pass
