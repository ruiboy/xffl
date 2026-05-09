# Score & Ladder Calculation — Implementation Plan

## Mental model

Two **inputs** determine when scores and ladders can be computed:
- `afl.match.data_status` — tracks AFL player stats completeness
- `ffl.club_match.data_status` — tracks FFL team submission and confirmation

Everything else (player scores, club match scores, match results, ladder standings) is **derived**
and stored in `drv_` columns. Those columns always reflect the current best-known calculation;
their data_status combination tells you whether to treat the value as provisional or final.

Two tiers of calculation:
- **Provisional** — `afl.match ∈ {partial, final}` AND `ffl.club_match ∈ {submitted, final}`
- **Final** — `afl.match = final` AND `ffl.club_match = final`

Only final results update the official ladder (`drv_` columns on `club_season`).
Provisional ladder is computed on-demand from current `drv_score` values — no dedicated event.

---

## Event model

### New events (add to `contracts/events/events.go`)

| Constant | Event string | Publisher | Trigger |
|----------|-------------|-----------|---------|
| `AflMatchFinalized` | `AFL.MatchFinalized` | AFL service | `afl.match.data_status → final` |
| `FflTeamSubmitted` | `FFL.TeamSubmitted` | FFL service | `ffl.club_match.data_status → submitted` |
| `FflTeamFinalized` | `FFL.TeamFinalized` | FFL service | `ffl.club_match.data_status → final` |
| `FflClubMatchScoreFinalized` | `FFL.ClubMatchScoreFinalized` | FFL service | AFL final + FFL team final (per club) |
| `FflMatchFinalized` | `FFL.MatchFinalized` | FFL service | Both clubs in FFL match finalized |

### Payloads

```go
AflMatchFinalizedPayload       { MatchID, SeasonID, RoundID int }
FflTeamSubmittedPayload        { ClubMatchID, MatchID, RoundID int }
FflTeamFinalizedPayload        { ClubMatchID, MatchID, RoundID int }
FflClubMatchScoreFinalizedPayload { ClubMatchID, MatchID int }
FflMatchFinalizedPayload       { MatchID, RoundID int }
```

---

## Event flow

```
AFL.PlayerMatchUpdated
  └─ FFL: recalc ffl.player_match.drv_score + ffl.club_match.drv_score (provisional, existing)

AFL.MatchFinalized
  ├─ AFL: set afl.match.drv_result (home_win/away_win/draw)
  ├─ AFL: recalc afl.club_season drv_ for both clubs (full season from scratch)
  └─ FFL: for each ffl.club_match in round:
           recalc provisional score
           if ffl.club_match.data_status = final → emit FFL.ClubMatchScoreFinalized

FFL.TeamSubmitted
  └─ FFL: recalc provisional score for this club_match

FFL.TeamFinalized
  ├─ FFL: recalc score for this club_match
  └─ if afl.match.data_status = final → emit FFL.ClubMatchScoreFinalized

FFL.ClubMatchScoreFinalized
  └─ FFL: SELECT COUNT(*) FROM ffl.club_match WHERE match_id=$1 AND data_status='final'
           if COUNT = 2 → emit FFL.MatchFinalized

FFL.MatchFinalized
  ├─ FFL: set ffl.match.drv_result (home_win/away_win/draw)
  └─ FFL: recalc ffl.club_season drv_ for both clubs (full season from scratch)
```

---

## Implementation plan

### Phase 1 — Contracts

- [x] Add `AflMatchFinalized` constant + `AflMatchFinalizedPayload` struct
- [x] Add `FflTeamSubmitted` constant + `FflTeamSubmittedPayload` struct
- [x] Add `FflTeamFinalized` constant + `FflTeamFinalizedPayload` struct
- [x] Add `FflClubMatchScoreFinalized` constant + `FflClubMatchScoreFinalizedPayload` struct
- [x] Add `FflMatchFinalized` constant + `FflMatchFinalizedPayload` struct

### Phase 2 — AFL service

#### Domain
- [x] `Match.DeriveResult() MatchResult` — derives result from StoredScore
- [x] `CalculateLadder(matches []Match) map[int]ClubSeason` — pure function; `ClubSeasonStats` was collapsed into `ClubSeason` (ID field added)

#### Repository
- [x] `UpdateMatchResult` + `FindFinalBySeasonID` on MatchRepository
- [x] `UpdateClubSeason` SQL + `Update` on ClubSeasonRepository

#### Application
- [x] `MarkMatchStatsFinal` extended — derives result, recalcs ladder, publishes `AFL.MatchFinalized` on `partial → final`

#### Data ops
- [x] `RecalculateAFLLadder(ctx, seasonID int) error`

### Phase 3 — FFL service

#### Domain
- [x] `Match.DeriveResult() MatchResult` — pure function using StoredScore
- [x] `CalculateLadder(matches []Match) map[int]ClubSeason` — `ClubSeasonStats` collapsed into `ClubSeason`; includes `ExtraPoints` field

#### Repository
- [x] `MatchRepository.UpdateResult(ctx, matchID int, result MatchResult) error`
- [x] `MatchRepository.FindFinalBySeasonID(ctx, seasonID int) ([]Match, error)`
- [x] `ClubSeasonRepository.Update(ctx, cs ClubSeason) error`
- [x] `ClubMatchRepository.CountFinalByMatchID(ctx, matchID int) (int, error)`

#### Application — publishing
- [x] `SetTeam` (was `ImportRoundTeams`) — publishes `FFL.TeamSubmitted` when data_status transitions to `submitted`
- [x] `MarkTeamFinal` — publishes `FFL.TeamFinalized` when data_status transitions to `final`

#### Application — handlers
- [x] `HandleAflMatchFinalized` — for each finalized ffl.club_match in round, emits `FFL.ClubMatchScoreFinalized`
- [x] `HandleFflTeamSubmitted` — removed; `SetTeam` calls `RecalculateClubMatchScore` directly after TX
- [x] `HandleFflTeamFinalized` — emits `FFL.ClubMatchScoreFinalized`
- [x] `HandleFflClubMatchScoreFinalized` — checks count; if both clubs final, emits `FFL.MatchFinalized`
- [x] `HandleFflMatchFinalized` — sets ffl.match.drv_result; recalcs FFL ladder for season
- [x] `RecalculateFflLadder(ctx, seasonID int) error` — walks all final FFL matches, rebuilds club_season drv_

#### Data ops
- [ ] `RecalculateFflScores(ctx, roundID int) error` — recalcs all FFL club_matches for AFL round (provisional and final)
- [ ] `ProvisionalLadder(ctx, seasonID int) ([]ClubSeasonStats, error)` — on-demand query; includes submitted+partial matches

### Phase 5 — Recalculate single club match score

#### Twirp RPC
- [x] `LookupPlayerMatch` refactored to `oneof {by_ids, by_season_round}` — single extensible endpoint
- [x] `PlayerMatchStats` proto gains `player_season_id` field (populated by `by_season_round` path)
- [x] AFL domain: `PlayerMatchRepository.FindByPlayerSeasonIDsAndRoundID` + SQL query
- [x] AFL Twirp handler: dispatches on `oneof` key

#### FFL use case
- [x] `PlayerLookup` port: `LookupPlayerMatchBySeasonRound(aflPSIDs, aflRoundID)` method
- [x] FFL rpc adapter: implements both `LookupPlayerMatch` (by_ids) and `LookupPlayerMatchBySeasonRound` (by_season_round)
- [x] `PlayerSeasonRepository.FindByIDs` batch fetch (SQL + domain + repo)
- [x] `EventRepos.ClubMatches` added for `clubMatchID → matchID → roundID → AFLRoundID` traversal
- [x] `RecalculateClubMatchScore` — two-path lookup: linked rows use AFL player_match_id (fast path); unlinked rows use `LookupPlayerMatchBySeasonRound` and establish the link as a side effect
- [x] `recalculateFFLClubMatchScore` GraphQL mutation wired up

#### Frontend
- [x] Recalculate button on each non-`no_data` row in FFL Teams tab

### Phase 4 — Data ops frontend

- [x] Add "Calculate" tab to data ops
- [x] Recalculate AFL ladder (per season) — `recalculateAFLLadder` mutation → `ScoreCommands.RecalculateAFLLadder`
- [x] Recalculate FFL ladder (per season) — `recalculateFFLLadder` mutation → `ScoreCommands.RecalculateFflLadder`
- [x] Mark FFL team final button — `markFFLTeamFinal` mutation → `DataOpsCommands.MarkTeamFinal` (in FFL Teams tab)
- [ ] Recalculate FFL scores for round — `RecalculateFflScores` (backend not yet implemented)
- [ ] Provisional ladder view — `ProvisionalLadder` (backend not yet implemented)

---

## TDD notes

- Domain functions (`Match.Result`, `CalculateClubSeasonStats`) are pure — unit test with table tests
- Handler functions — use in-memory dispatcher; test event publish/subscribe chains
- Repo methods — integration tests against real DB (see `ai/architecture/testing.md`)
- Data ops commands — integration tests; seed data with known final statuses, assert drv_ columns