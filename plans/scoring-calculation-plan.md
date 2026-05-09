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

- [ ] Add `AflMatchFinalized` constant + `AflMatchFinalizedPayload` struct
- [ ] Add `FflTeamSubmitted` constant + `FflTeamSubmittedPayload` struct
- [ ] Add `FflTeamFinalized` constant + `FflTeamFinalizedPayload` struct
- [ ] Add `FflClubMatchScoreFinalized` constant + `FflClubMatchScoreFinalizedPayload` struct
- [ ] Add `FflMatchFinalized` constant + `FflMatchFinalizedPayload` struct

### Phase 2 — AFL service

#### Domain
- [ ] `Match.Result() MatchResult` — pure function deriving `home_win/away_win/draw/no_result` from club match scores
- [ ] `CalculateClubSeasonStats(matches []Match) ClubSeasonStats` — pure function; folds all final matches into played/won/lost/drawn/for/against/premiership_points

#### Repository
- [ ] `UpdateMatchResult(ctx, matchID int, result string) error`
- [ ] `UpdateClubSeason(ctx, clubSeasonID int, stats ClubSeasonStats) error`
- [ ] `GetFinalMatchesForSeason(ctx, seasonID int) ([]Match, error)` — returns matches where data_status = final

#### Application
- [ ] `SetMatchDataStatus(ctx, matchID, status)` — already exists; extend to publish `AFL.MatchFinalized` on `partial → final` transition
- [ ] `HandleAflMatchFinalized(ctx, payload)` — sets drv_result, recalcs AFL ladder for season

#### Data ops
- [ ] `RecalculateAFLLadder(ctx, seasonID int) error` — walks all final matches, rebuilds club_season drv_ (manual trigger)

### Phase 3 — FFL service

#### Domain
- [ ] `Match.Result() MatchResult` — pure function; same logic as AFL Match.Result()
- [ ] `CalculateClubSeasonStats(matches []Match) ClubSeasonStats` — same shape as AFL; includes extra_points field

#### Repository
- [ ] `UpdateMatchResult(ctx, matchID int, result string) error`
- [ ] `UpdateClubSeason(ctx, clubSeasonID int, stats ClubSeasonStats) error`
- [ ] `CountFinalizedClubMatches(ctx, matchID int) (int, error)` — count where data_status = final
- [ ] `GetFinalMatchesForSeason(ctx, seasonID int) ([]Match, error)`
- [ ] `GetClubMatchesForAFLRound(ctx, roundID int) ([]ClubMatch, error)` — for provisional recalc

#### Application — publishing
- [ ] `SubmitTeam(...)` — publish `FFL.TeamSubmitted` when data_status transitions to `submitted`
- [ ] `FinalizeTeam(...)` — publish `FFL.TeamFinalized` when data_status transitions to `final`

#### Application — handlers
- [ ] `HandleAflMatchFinalized` — recalcs provisional FFL scores for round; for each finalized ffl.club_match emits `FFL.ClubMatchScoreFinalized`
- [ ] `HandleFflTeamSubmitted` — recalcs provisional score for this club_match
- [ ] `HandleFflTeamFinalized` — recalcs score; if afl.match = final, emits `FFL.ClubMatchScoreFinalized`
- [ ] `HandleFflClubMatchScoreFinalized` — checks count; if both clubs final, emits `FFL.MatchFinalized`
- [ ] `HandleFflMatchFinalized` — sets ffl.match.drv_result; recalcs FFL ladder for season

#### Data ops
- [ ] `RecalculateFflScores(ctx, roundID int) error` — recalcs all FFL club_matches for AFL round (provisional and final)
- [ ] `RecalculateFflLadder(ctx, seasonID int) error` — walks all final FFL matches, rebuilds club_season drv_
- [ ] `ProvisionalLadder(ctx, seasonID int) ([]ClubSeasonStats, error)` — on-demand query; includes submitted+partial matches

### Phase 4 — Data ops frontend

- [ ] Add "Calculate" tab to data ops
- [ ] Recalculate AFL ladder (per season) — calls `RecalculateAFLLadder`
- [ ] Recalculate FFL scores for round — calls `RecalculateFflScores`
- [ ] Recalculate FFL ladder (per season) — calls `RecalculateFflLadder`
- [ ] Provisional ladder view — calls `ProvisionalLadder`

---

## TDD notes

- Domain functions (`Match.Result`, `CalculateClubSeasonStats`) are pure — unit test with table tests
- Handler functions — use in-memory dispatcher; test event publish/subscribe chains
- Repo methods — integration tests against real DB (see `ai/architecture/testing.md`)
- Data ops commands — integration tests; seed data with known final statuses, assert drv_ columns