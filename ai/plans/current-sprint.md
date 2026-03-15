# Current Sprint

**Sprint goal:** Complete Phase 2 — AFL service with sqlc

## Tasks

### sqlc migration
- [ ] Set up sqlc config (`services/afl/sqlc.yaml`)
- [ ] Write SQL query files for all repositories
- [ ] Generate sqlc code, wire into service
- [ ] Remove raw pgx repository implementations

### Mutations
- [ ] Application layer — `UpdatePlayerMatch` use case with `DB.WithTx`
- [ ] Infrastructure layer — write repository methods (Create/Update)
- [ ] Implement `updateAFLPlayerMatch` resolver

### Tests
- [ ] Mutation integration tests
- [ ] Error case tests (not found, conflicts)
