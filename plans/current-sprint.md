# Current Sprint

**Sprint goal:** Phase 11 — FFL Event Integration

Wire up cross-service event flow: AFL stat updates automatically trigger FFL fantasy score recalculation.

## Tasks

### 1. Contract: extend event payload
- [ ] Add `RoundID int` to `PlayerMatchUpdatedPayload` in `contracts/events/events.go`

### 2. AFL: publish event on stat update
- [ ] Add `events.Dispatcher` to AFL `Commands` (new dependency)
- [ ] Publish `AFL.PlayerMatchUpdated` after `UpdatePlayerMatch` succeeds
- [ ] Wire PG dispatcher in `services/afl/cmd/main.go`, start `Listen` goroutine
- [ ] Unit/integration test: verify event is published with correct payload

### 3. FFL schema: round correlation
- [ ] Add `afl_round_id INTEGER` column to `ffl.round` in `dev/postgres/init/02_ffl_schema.sql`
- [ ] Populate `afl_round_id` in seed data
- [ ] Add field to FFL `domain.Round`

### 4. FFL infra: new queries
- [ ] SQLC query: `FindRoundByAFLRoundID` — find FFL round by `afl_round_id`
- [ ] SQLC query: `FindPlayerMatchByPlayerSeasonAndRound` — join `player_match → club_match → match` to find player_match by `(player_season_id, round_id)`
- [ ] SQLC query: `FindPlayerSeasonsByAFLPlayerSeasonID` — find all FFL player_seasons for a given AFL player_season
- [ ] Add repository methods + domain interfaces for the above
- [ ] `sqlc generate` + update `repository.go`

### 5. FFL application: event handler
- [ ] New `HandlePlayerMatchUpdated(ctx, payload)` method on `Commands`
  - Decode `PlayerMatchUpdatedPayload`
  - Find FFL round by `afl_round_id`
  - Find FFL player_seasons by `afl_player_season_id`
  - For each: find FFL player_match by `(player_season_id, round_id)` via join query
  - Call `CalculateFantasyScore` with AFL stats
  - Set `afl_player_match_id` on the FFL player_match
  - Publish `FFL.FantasyScoreCalculated`
- [ ] Add `events.Dispatcher` dependency to FFL `Commands`

### 6. FFL main.go: wire dispatcher
- [ ] Wire PG dispatcher in `services/ffl/cmd/main.go`
- [ ] Subscribe `HandlePlayerMatchUpdated` to `AFL.PlayerMatchUpdated`
- [ ] Start `Listen` goroutine

### 7. Integration tests
- [ ] Test: AFL publishes event → FFL handler scores the correct player_match
- [ ] Test: event for player not in any FFL squad is silently ignored
- [ ] Test: multiple FFL clubs with same AFL player all get scored
