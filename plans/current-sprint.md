# Current Sprint

**Sprint goal:** Phase 5 — FFL Service (Fantasy Football League backend)

Build the FFL service following the same clean architecture patterns as the AFL service: domain entities with business logic, sqlc for data access, gqlgen for GraphQL, and integration tests against a real database. The FFL service mirrors the AFL service's structure (league/season/round/match/club/player hierarchy) but adds fantasy scoring. Event subscription (AFL→FFL) is deferred to a later task.

## Tasks

### 1. Domain layer
- [x] Core entities: League, Season, Round, Match, ClubSeason, ClubMatch, Player, PlayerSeason, PlayerMatch
- [x] Position-based scoring: goals(5/goal), kicks(1/kick), handballs(1/handball), marks(2/mark), tackles(4/tackle), hitouts(1/hitout), star(5/goal+1/kick+1/handball+2/mark+4/tackle)
- [x] PlayerMatch fields: position, status, backup_positions (nullable string), interchange_position (nullable string), score
- [x] Domain methods: PlayerMatch.CalculateScore(aflStats), ClubMatch.Score(), ClubSeason.Percentage()
- [x] Bench/sub logic: sub only when starter DNPs, interchange auto-swaps if bench outscores starter
- [x] Repository interfaces on each entity
- [x] Unit tests for scoring by position, percentage, bench substitution rules

### 2. Application layer
- [x] Queries: clubs, players, seasons, rounds, matches, ladder, player matches
- [x] Commands: ManagePlayers (CRUD), CalculateFantasyScore
- [x] TxManager interface (same pattern as AFL)

### 3. Infrastructure — sqlc + Postgres
- [x] SQL query files for all entities (ffl schema)
- [x] sqlc config and code generation
- [x] Repository implementations mapping sqlcgen → domain
- [x] Transaction manager implementation

### 4. Interface — GraphQL
- [x] Schema: queries (clubs, players, seasons, ladder, matches) + mutations (CRUD players, update player match)
- [x] gqlgen config and code generation
- [x] Query/mutation resolvers + field resolvers for nested types
- [x] Converter layer (domain ↔ GraphQL)

### 5. Service wiring
- [ ] cmd/main.go: DB pool → repos → queries/commands → resolver → HTTP server (port 8081)
- [ ] go.mod with pgx, gqlgen dependencies
- [ ] Health endpoint

### 6. Gateway routing
- [ ] Add FFL service proxy to gateway (route FFL queries to :8081)
- [ ] Update run-all in justfile to include FFL service

### 7. Integration tests
- [ ] GraphQL integration tests (queries + mutations) against real DB
- [ ] Test helpers: seed data, server setup, query execution
- [ ] Fantasy score calculation test

### 8. Validate end-to-end
- [ ] `just dev-up && just dev-seed` loads FFL data
- [ ] `just run-ffl` starts and serves GraphQL playground
- [ ] Gateway proxies FFL queries correctly
- [ ] All tests pass

## Design constraint: event integration is next
The AFL service already publishes `AFL.PlayerMatchUpdated` events. The next sprint will wire up the FFL service to subscribe to these events and auto-calculate fantasy scores. This sprint should ensure:
- `PlayerMatch.CalculateScore(aflStats)` is a pure domain function, callable from a future event handler
- The application layer command for calculating scores doesn't assume where the AFL stats come from

## Out of scope (deferred)
- Event subscription (AFL.PlayerMatchUpdated → FFL.FantasyScoreCalculated) — Phase 7
- Event publishing (FFL.FantasyScoreCalculated) — Phase 7
- Frontend FFL views — Phase 6
- Draft/trade mechanics — future phase
