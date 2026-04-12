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

Frontend-driven: frontend calls AFL `liveRound`, then maps to FFL round by `afl_round_id` client-side. No FFL → AFL service dependency introduced yet.

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
- [ ] Create `useAflState.ts`: manages `xffl_afl` JSON cookie `{ seasonId, roundId, roundStatus }`
- [ ] Refactor `useFflState.ts`: replace flat string refs + `setCurrentSeason` with `xffl_ffl` JSON cookie, same shape
- [ ] On page load: call `liveRound` query for the relevant service, update cookie; trust stale value on failure

### 8. Frontend: RoundNav
- [ ] Feed `liveRoundId` from cookie (not URL)
- [ ] Pass `roundStatus` alongside `liveRoundId`; apply distinct ring style for `Open` vs `Closed`

## Revisit at end of sprint

- [ ] `adelaideLoc` in `domain/round.go` — not fully convinced this belongs in domain. Consider whether timezone config belongs in application layer or as a domain constant. Revisit once FFL side is implemented and the full pattern is visible.
