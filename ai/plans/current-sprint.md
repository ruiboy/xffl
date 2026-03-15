# Current Sprint

**Sprint goal:** Complete Phase 1 — shared packages, contracts, and tooling

## Decisions
- [x] Resolve ADR-009 — sqlc + pgx with thin helper layer (accepted 2026-03-15)
- [x] Resolve ADR-004 — PG LISTEN/NOTIFY as event transport (accepted 2026-03-15)

## Tasks
- [ ] `shared/database/` — DB connection helper (WithTx, Queries, MapPgError)
- [ ] Migration tooling — replace raw SQL files
- [ ] `contracts/events/` — shared event type definitions (PlayerMatchUpdated, FantasyScoreCalculated)
- [ ] `shared/events/` — EventDispatcher interface + PG LISTEN/NOTIFY implementation
- [ ] `shared/events/memory/` — in-memory dispatcher for testing