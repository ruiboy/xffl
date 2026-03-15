# Current Sprint

**Sprint goal:** Complete Phase 1 — shared packages, contracts, and tooling

## Decisions
- [ ] Resolve ADR-009 — persistence layer (raw pgx vs sqlc+pgx vs GORM); define transaction boundary pattern
- [ ] Resolve ADR-004 — event transport (PG LISTEN/NOTIFY vs NATS); update ADR with decision

## Tasks
- [ ] `shared/database/` — DB connection helper (depends on ADR-009)
- [ ] Migration tooling — replace raw SQL files
- [ ] `contracts/events/` — shared event type definitions (PlayerMatchUpdated, FantasyScoreCalculated)
- [ ] `shared/events/` — EventDispatcher interface + implementation (depends on ADR-004); update ADR-004
- [ ] `shared/events/memory/` — in-memory dispatcher for testing