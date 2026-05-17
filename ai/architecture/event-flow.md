# Event Flow

## AFL Events

**Published:**
- `AFL.PlayerMatchUpdated` — fired when a player's match stats change. Payload carries full stats (kicks, handballs, marks, hitouts, tackles, goals, behinds). No status field — participation status is carried exclusively by `AFL.MatchUpdated`.
- `AFL.MatchUpdated` — fired on match status transitions only (`no_data → partial`, `partial → final`). Carries `match_status` and `PlayerSeasonIDStatusMap`: a map of `afl_player_season_id → status`. On `partial`: every player currently with stats → `"playing"`. On `final`: players with stats → `"played"`; players in both club squads with no stats → `"dnp"`. AFL derives match result and recalculates the AFL ladder internally on `final` — no cross-service event is emitted for this.

---

## FFL Events

**Subscribes to:**
- `AFL.PlayerMatchUpdated` → links `afl_player_match_id` if not yet set; recalculates fantasy score for the player and club_match total; if both axes are already final (`AllAFLStatusesFinal` + `data_status = final`) recalculates FFL ladder (post-final stat correction cascade).
- `AFL.MatchUpdated` → applies `PlayerSeasonIDStatusMap` to set `drv_afl_status` on matching FFL player_matches (players not in the map belong to a different AFL match — ignore); recalculates affected club_match scores; if `ffl.club_match.data_status = final` AND `AllAFLStatusesFinal` → emits `FFL.ClubMatchScoreFinalized`.

**Publishes:**
- `FFL.ClubMatchUpdated` — fired on any change to a club's team for a round: initial submission, correction, subs declared, or finalization. Carries `data_status` (`submitted` | `final`) and a full snapshot of all player_match positions and statuses (`map[player_match_id]FflPlayerMatchInfo`).
- `FFL.PlayerMatchUpdated` — fired after each player's fantasy score is calculated. Carries `player_match_id`, `club_match_id`, and `score`.
- `FFL.ClubMatchScoreFinalized` — fired when a single club's score is locked: `ffl.club_match.data_status = final` AND `AllAFLStatusesFinal`. Fires independently per club.
- `FFL.MatchScoreFinalized` — fired when both clubs in an FFL match have emitted `FFL.ClubMatchScoreFinalized`. Triggers `ffl.match.drv_result` derivation and FFL ladder recalculation.

**Subscribes to own events:**
- `FFL.ClubMatchUpdated` → recalculates score; if `data_status = final` AND `AllAFLStatusesFinal` → emits `FFL.ClubMatchScoreFinalized`.
- `FFL.ClubMatchScoreFinalized` → if both clubs in the FFL match are now finalized → emits `FFL.MatchScoreFinalized`.
- `FFL.MatchScoreFinalized` → derives and persists `ffl.match.drv_result`; recalculates FFL ladder for the season.

---

## Search subscriptions

The Search service indexes entities as they change by subscribing to domain events from AFL and FFL.

- `AFL.PlayerMatchUpdated` → index player match document
- `AFL.PlayerSeasonUpdated` → index player season document *(trigger not yet wired — Phase 23)*
- `FFL.PlayerMatchUpdated` → index fantasy score document

---

## Data Calculation Flow

### Mental Model

Two **inputs** determine when scores and ladders can be computed:
- **AFL match** data status — tracks player stats completeness
- **FFL ClubMatch** data status — tracks team submission and confirmation

Everything else (PlayerMatch scores, ClubMatch scores, Match results, ClubSeason ladder standings) is **derived**. Derived fields always reflect the current best-known calculation; the data status combination tells you whether to treat the value as provisional or final.

Two tiers of calculation:
- **Provisional** — AFL match ∈ {partial, final} AND FFL ClubMatch = submitted
- **Final** — AFL match = final AND FFL ClubMatch = final AND `AllAFLStatusesFinal`

Only final results update the official ladder (ClubSeason derived fields).
Provisional ladder is computed on-demand from current derived scores — no dedicated event.

### AllAFLStatusesFinal

A shared domain check used in three handler paths:

> `AllAFLStatusesFinal(clubMatchID) bool` — returns true when every `ffl.player_match` in the club_match has `drv_afl_status ∈ {played, dnp}` (none are null or playing).

FFL uses this to infer AFL finality from its own data, without a cross-service call. It is true only once every AFL match that any player in the FFL team participates in has been finalised and processed.

### Event Chain

```
AFL.PlayerMatchUpdated (stats only — no status)
  └─ FFL: link afl_player_match_id if unset; recalculate score
          └─ if data_status = final AND AllAFLStatusesFinal → RecalculateFflLadder

AFL.MatchUpdated (partial)
  ├─ AFL: [no action — internal state already updated]
  └─ FFL: set drv_afl_status = playing for players in map; recalculate affected scores

AFL.MatchUpdated (final)
  ├─ AFL: derive match result; recalculate AFL ladder [internal — no cross-service event]
  └─ FFL: set drv_afl_status from map (played / dnp); recalculate affected scores
          └─ if ffl.club_match.data_status = final AND AllAFLStatusesFinal
              → FFL.ClubMatchScoreFinalized

FFL.ClubMatchUpdated (data_status = submitted — team change, correction, or subs declared)
  └─ FFL: recalculate provisional score

FFL.ClubMatchUpdated (data_status = final)
  └─ FFL: recalculate score
          └─ if AllAFLStatusesFinal → FFL.ClubMatchScoreFinalized

FFL.ClubMatchScoreFinalized               (fires per club, independently)
  └─ FFL: if both clubs in the FFL match are now finalized → FFL.MatchScoreFinalized

FFL.MatchScoreFinalized
  └─ FFL: derive match result; recalculate FFL ladder
```

Ladder recalculation (both AFL and FFL): the entire season is recalculated from scratch on each trigger — simpler and drift-free given bounded season length (~22 rounds).

AFL ladder: recalculated internally in `MarkMatchStatsFinal` on `data_status → final`. Not re-triggered by post-final stat corrections (AFL match results do not change from stat corrections).

FFL ladder: recalculated on `FFL.MatchScoreFinalized` and also directly when `AFL.PlayerMatchUpdated` arrives and both axes are already final (post-final stat correction auto-cascade).

Provisional ladder is computed on-demand from current derived scores — no dedicated event.