# Current Sprint — Phase 12: Live Round

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

Two-step frontend query: (1) `aflLiveRound` → AFL round ID + status; (2) `fflRoundByAflRound(aflRoundId)` → FFL round (nullable). If null, show "Cannot determine round to display. Consult your admin." No `fflLatestRound` on home page — it returned the last round by insertion order, which could be a future round.

## Tasks

### 1. Shared: Clock interface
- [x] Define `Clock` interface in `shared/clock/`
- [x] `RealClock` implementation wrapping `time.Now()`
- [x] `FixedClock` implementation for tests
- [x] `CLOCK_OVERRIDE` env var support: if set at startup, wire `FixedClock`; otherwise `RealClock`

### 2. AFL: RoundStatus type + LiveRound use case
- [x] Add `RoundStatus` type (`Open` / `Closed`) to AFL domain
- [x] Add `LiveRoundResult` struct: `Round`, `Status RoundStatus`
- [x] SQLC query: fetch all rounds for a season with `MIN(match.start_dt)` and `MAX(match.start_dt)`
- [x] `LiveRound(ctx) LiveRoundResult` use case: loads round+match bounds, computes midnight windows in Australia/Adelaide, returns Open if window active, Closed for most recently completed
- [x] Wire `Clock` into AFL `Queries`

### 3. AFL: GraphQL
- [x] Add `aflLiveRound: AFLLiveRound` query to AFL schema
- [x] Resolver calls `queries.LiveRound(ctx)`

### 4 & 5. FFL: expose aflRoundId on FFLRound
- [x] Add `aflRoundId: ID` to `FFLRound` in FFL GraphQL schema
- [x] Populate `AflRoundID` in `convertRound`
- No clock, no use case, no `fflLiveRound` query needed — frontend maps AFL→FFL round client-side using `aflRoundId`

### 6. E2e test seed data
- [x] Add AFL Round 3 with `start_dt = 2026-01-15 14:10:00+10:30` to e2e AFL seed
- [x] Add FFL Round 3 with `afl_round_id` pointing to AFL Round 3 to e2e FFL seed
- [x] `CLOCK_OVERRIDE=2026-01-15T10:00:00+10:30` wired into both service commands in `playwright.config.ts`

### 7. Frontend: state composables
- [x] Create `useAflState.ts`: manages `xffl_afl` JSON cookie `{ seasonId, roundId, roundStatus }`
- [x] Refactor `useFflState.ts`: replace flat string refs + `setCurrentSeason` with `xffl_ffl` JSON cookie, same shape
- [x] On page load: call `liveRound` query for the relevant service, update cookie; trust stale value on failure

### 8. Frontend: RoundNav
- [x] Feed `liveRoundId` from cookie (not URL)
- [x] Pass `roundStatus` alongside `liveRoundId`; apply distinct ring style for `Open` vs `Closed`

## Revisit at end of sprint

- [x] `adelaideLoc` in `domain/round.go` — resolved: moved to `application/live_round.go` as part of the FindNeighbours refactor. Timezone logic lives in the application layer alongside the `LiveRound` use case.
