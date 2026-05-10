# Event Flow

## AFL Events

**Published:**
- `AFL.PlayerMatchUpdated` — fired when a player's match stats change. Payload carries full stats (kicks, handballs, marks, hitouts, tackles, goals, behinds).
- `AFL.MatchFinalized` — fired when `afl.match.data_status` transitions to `final`. Triggers AFL match result derivation and AFL ladder recalculation. Also signals the FFL service to recalculate affected club match scores.

---

## FFL Events

**Subscribes to:**
- `AFL.PlayerMatchUpdated` → incremental provisional score update for the affected player and club match.
- `AFL.MatchFinalized` → recalculate provisional/final FFL scores for all club matches in the round.

**Publishes:**
- `FFL.FantasyScoreCalculated` — carries the calculated score and the AFL PlayerMatch ID it was derived from.
- `FFL.TeamSubmitted` — fired when `ffl.club_match.data_status → submitted`. Triggers provisional score calculation.
- `FFL.TeamFinalized` — fired when `ffl.club_match.data_status → final`. Triggers final score calculation if AFL is also final.
- `FFL.ClubMatchScoreFinalized` — fired when a single club's score is locked (AFL final + FFL team final). Triggers check for full match finalization.
- `FFL.MatchFinalized` — fired when both clubs in an FFL match have finalized. Triggers `ffl.match.drv_result` derivation and ladder recalculation. Symmetric with `AFL.MatchFinalized`.

---

## Search subscriptions

The Search service indexes entities as they change by subscribing to domain events from AFL and FFL.

- `AFL.PlayerMatchUpdated` → index player match document
- `AFL.PlayerSeasonUpdated` → index player season document *(trigger not yet wired — Phase 23)*
- `FFL.FantasyScoreCalculated` → index fantasy score document

---

## Data Calculation Flow

### Mental Model

Two **inputs** determine when scores and ladders can be computed:
- **AFL Match** data status — tracks player stats completeness
- **FFL ClubMatch** data status — tracks team submission and confirmation

Everything else (PlayerMatch scores, ClubMatch scores, Match results, ClubSeason ladder standings) is **derived**. Derived fields always reflect the current best-known calculation; the data status combination tells you whether to treat the value as provisional or final.

Two tiers of calculation:
- **Provisional** — AFL Match ∈ {partial, final} AND FFL ClubMatch ∈ {submitted, final}
- **Final** — AFL Match = final AND FFL ClubMatch = final

Only final results update the official ladder (ClubSeason derived fields).
Provisional ladder is computed on-demand from current derived scores — no dedicated event.

### Event Chain

The following events chain AFL and FFL score and ladder derivation. Both services react to each other's finalization events.

```
AFL.PlayerMatchUpdated
  └─ FFL: update provisional player and club match scores

AFL.MatchFinalized
  ├─ AFL: derive match result; recalculate AFL ladder
  └─ FFL: recalculate scores for all FFL club matches in the round
          └─ if ffl.club_match.data_status = final → FFL.ClubMatchScoreFinalized

FFL.TeamSubmitted
  └─ FFL: recalculate provisional score for this club match

FFL.TeamFinalized
  └─ FFL: recalculate score for this club match
          └─ if afl.match.data_status = final → FFL.ClubMatchScoreFinalized

FFL.ClubMatchScoreFinalized               (fires per club, independently)
  └─ FFL: if both clubs in the FFL match are now final → FFL.MatchFinalized

FFL.MatchFinalized
  └─ FFL: derive match result; recalculate FFL ladder
```

Ladder recalculation (both AFL and FFL): the entire season is recalculated from scratch on each trigger — simpler and drift-free given bounded season length (~22 rounds).

Provisional ladder is computed on-demand from current derived scores — no dedicated event.