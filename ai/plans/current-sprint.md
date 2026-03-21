# Current Sprint

**Sprint goal:** Complete Phase 2 — AFL service with sqlc

## Tasks

### sqlc migration
- [x] Set up sqlc config (`services/afl/sqlc.yaml`)
- [x] Write SQL query files for all repositories
- [x] Generate sqlc code, wire into service
- [x] Remove raw pgx repository implementations

### Mutations
- [x] Application layer — `UpdatePlayerMatch` use case with `DB.WithTx`
- [x] Infrastructure layer — write repository methods (Create/Update)
- [x] Implement `updateAFLPlayerMatch` resolver

### Tests
- [x] Mutation integration tests
- [x] Error case tests (not found, conflicts)
