# Current Sprint

**Sprint goal:** Phase 11 — FFL Event Integration

Wire up cross-service event flow: AFL stat updates automatically trigger FFL fantasy score recalculation.

## Tasks

### 1. Contract: extend event payload
- [x] Add `RoundID int` to `PlayerMatchUpdatedPayload` in `contracts/events/events.go`

### 2. AFL: publish event on stat update
- [x] Add `events.Dispatcher` to AFL `Commands` (new dependency)
- [x] Publish `AFL.PlayerMatchUpdated` after `UpdatePlayerMatch` succeeds
- [x] Wire PG dispatcher in `services/afl/cmd/main.go`, start `Listen` goroutine
- [x] Add `FindRoundIDByClubMatchID` SQLC query + domain interface + repo method

### 3. FFL schema: round correlation
- [x] Add `afl_round_id INTEGER` column to `ffl.round` in `dev/postgres/init/02_ffl_schema.sql`
- [x] Populate `afl_round_id` in seed data
- [x] Add field to FFL `domain.Round`

### 4. FFL infra: new queries
- [x] SQLC query: `FindRoundByAFLRoundID` — find FFL round by `afl_round_id`
- [x] SQLC query: `FindPlayerMatchByPlayerSeasonAndRound` — join `player_match → club_match → match` to find player_match by `(player_season_id, round_id)`
- [x] SQLC query: `FindPlayerSeasonsByAFLPlayerSeasonID` — find all FFL player_seasons for a given AFL player_season
- [x] SQLC query: `UpdateAFLPlayerMatchID` — set the AFL link on player_match
- [x] Add repository methods + domain interfaces for the above
- [x] `sqlc generate` + update `repository.go`

### 5. FFL application: event handler
- [x] New `HandlePlayerMatchUpdated(ctx, payload)` method on `Commands`
  - Decode `PlayerMatchUpdatedPayload`
  - Find FFL round by `afl_round_id`
  - Find FFL player_seasons by `afl_player_season_id`
  - For each: find FFL player_match by `(player_season_id, round_id)` via join query
  - Call `CalculateFantasyScore` with AFL stats
  - Set `afl_player_match_id` on the FFL player_match
  - Publish `FFL.FantasyScoreCalculated`
- [x] Add `events.Dispatcher` dependency to FFL `Commands`

### 6. FFL main.go: wire dispatcher
- [x] Wire PG dispatcher in `services/ffl/cmd/main.go`
- [x] Subscribe `HandlePlayerMatchUpdated` to `AFL.PlayerMatchUpdated`
- [x] Start `Listen` goroutine

### 7. Integration tests
- [x] Test: AFL publishes event → FFL handler scores the correct player_match
- [x] Test: event for player not in any FFL squad is silently ignored
- [x] Test: multiple FFL clubs with same AFL player all get scored
