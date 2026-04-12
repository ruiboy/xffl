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

---

# Phase 12 — Live Round

**Sprint goal:** Compute and expose a "live round" across AFL and FFL services, drive the round nav default and indicator from it, and make the whole thing testable without real-time dependency.

## Design decisions

- **Live window**: midnight (Australia/Adelaide) before first game of round → midnight after last game of round. Derived from `MIN`/`MAX` of `afl.match.start_dt` per round.
- **Fallback**: if no round window is currently open, return the most recently completed round.
- **`RoundStatus`**: `Open` (window active) / `Closed` (fallback — most recently completed).
- **Domain owns logic**: DB query returns round + match bounds; midnight-boundary calculation and Open/Closed determination live in Go domain code, not SQL.
- **Injectable `Clock` interface**: wired at startup. Overridden via `CLOCK_OVERRIDE=<RFC3339>` env var — used by e2e tests to fix time without exposing time-travel on the API.
- **Cookies**: two JSON cookies — `xffl_afl` and `xffl_ffl` — each `{ seasonId, roundId, roundStatus }`. Refreshed on page load. Trust stale value if query fails.
- **Nav indicator**: `liveRoundId` in RoundNav is fed from cookie (not URL). Independent of the currently browsed round. `Open` vs `Closed` status can drive different ring styles.

## Decision: FFL live round mapping

Frontend-driven: frontend calls AFL `liveRound`, then maps to FFL round by `afl_round_id` client-side. No FFL → AFL service dependency introduced yet.

## Tasks

### 1. Shared: Clock interface
- [ ] Define `Clock` interface in `shared/` (or per-service if no shared location is appropriate)
- [ ] `RealClock` implementation wrapping `time.Now()`
- [ ] `FixedClock` implementation for tests
- [ ] `CLOCK_OVERRIDE` env var support: if set at startup, wire `FixedClock`; otherwise `RealClock`

### 2. AFL: RoundStatus type + LiveRound use case
- [ ] Add `RoundStatus` type (`Open` / `Closed`) to AFL domain
- [ ] Add `LiveRoundResult` struct: `Round *Round`, `Status RoundStatus`
- [ ] SQLC query: fetch all rounds for a season with `MIN(match.start_dt)` and `MAX(match.start_dt)`
- [ ] `LiveRound(ctx, asOf time.Time) (*LiveRoundResult, error)` use case:
  - Load round+match bounds for current season
  - For each round: compute window in Australia/Adelaide timezone
  - Return first round whose window contains `asOf`; if none, return most recently completed with `Closed`
- [ ] Wire `Clock` into AFL `Commands`

### 3. AFL: GraphQL
- [ ] Add `liveRound: LiveRoundResult` query to AFL schema
- [ ] Resolver calls `commands.LiveRound(ctx, clock.Now())`

### 4. FFL: LiveRound use case
- [ ] Add `RoundStatus` type + `LiveRoundResult` to FFL domain
- [ ] `LiveRound` use case: find FFL round matching live AFL round via `afl_round_id`; fall back to most recently completed FFL round
- [ ] Wire `Clock` into FFL `Commands`

### 5. FFL: GraphQL
- [ ] Add `liveRound: LiveRoundResult` query to FFL schema
- [ ] Resolver calls `commands.LiveRound(ctx, clock.Now())`

### 6. E2e test seed data
- [ ] Add a dedicated e2e AFL round with `afl.match.start_dt` values around a fixed date (e.g. `2026-01-15`)
- [ ] Add matching FFL round with `afl_round_id` pointing to above
- [ ] Configure e2e harness to start services with `CLOCK_OVERRIDE=2026-01-15T10:00:00+10:30`

### 7. Frontend: state composables
- [ ] Create `useAflState.ts`: manages `xffl_afl` JSON cookie `{ seasonId, roundId, roundStatus }`
- [ ] Refactor `useFflState.ts`: replace flat string refs + `setCurrentSeason` with `xffl_ffl` JSON cookie, same shape
- [ ] On page load: call `liveRound` query for the relevant service, update cookie; trust stale value on failure

### 8. Frontend: RoundNav
- [ ] Feed `liveRoundId` from cookie (not URL)
- [ ] Pass `roundStatus` alongside `liveRoundId`; apply distinct ring style for `Open` vs `Closed`
