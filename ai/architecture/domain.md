# Ubiquitous Language

Shared vocabulary for the xffl codebase. Every entity, value, and rule listed here should be used consistently in code, tests, docs, and conversation.

## Shared Entities

These entities appear in both AFL and FFL bounded contexts. They share the same names but belong to separate schemas (`afl.*`, `ffl.*`) — no cross-service imports.

| Term | Meaning |
|------|---------|
| **League** | A competition (e.g. the AFL, a fantasy league). Container for seasons. |
| **Season** | One year of a league. Contains rounds. |
| **Round** | A set of matches within a season. May span multiple days. |
| **Match** | A game between two clubs. Composed of a Home and Away ClubMatch. |
| **Club** | A team. |
| **ClubSeason** | A club's record for one season (played, won, lost, drawn, for, against). Used to build the ladder. |
| **ClubMatch** | One side of a match — a club's performance in that game. |
| **Player** | An individual athlete. Exists independently — club association is through PlayerSeason. |
| **PlayerSeason** | A player's registration with a club for a season (via ClubSeason). This is where the player–club relationship lives. Includes `from_round_id` / `to_round_id` to track when a player joined or left a club during the season (null = start/end of season). |
| **PlayerMatch** | A player's involvement in a single match. Fields differ by context (see below). |

### Derived fields

Columns prefixed `drv_` in the database are **derived** (computed from other data). Examples: `drv_score`, `drv_played`, `drv_premiership_points`. Domain entities drop the prefix — `ClubSeason.Won` maps to `drv_won`.

---

## AFL Context

Real-world Australian Football League data.

### Scoring

| Term | Meaning |
|------|---------|
| **Goal** | Worth 6 points. Kicked through the tall posts. |
| **Behind** | Worth 1 point. Through the short posts, or off hands/body. |
| **Rushed behind** | A behind conceded by the defending team (not from a kick). Counted on ClubMatch, not PlayerMatch. |
| **Score** | `Goals × 6 + Behinds`. For a club: sum of player scores + rushed behinds. |

### Player statistics

| Term | Meaning |
|------|---------|
| **Kick** | A foot-pass or shot on goal. |
| **Handball** | A hand-pass (punch the ball from one hand). |
| **Disposal** | Kick + handball. A derived stat. |
| **Mark** | A clean catch from a kick of 15+ metres. |
| **Tackle** | Wrapping up an opponent with the ball. |
| **Hitout** | Tapping the ball at a ruck contest (centre bounce, boundary throw-in). |
| **Goal** | Also a player stat — number of goals kicked by the player in a match. |
| **Behind** | Also a player stat — number of behinds scored by the player in a match. |
| **Minutes played** | Time on ground during the match. |

### Ladder

| Term | Meaning |
|------|---------|
| **Premiership points** | Win = 4, draw = 2, loss = 0. |
| **For / Against** | Total points scored / conceded across the season. |
| **Percentage** | `For ÷ Against × 100`. Tiebreaker on the ladder. |

### Match data status

Tracks completeness of player stats for a match. Key input for scoring calculations.

| Status | Meaning |
|--------|---------|
| `no_data` | No stats imported yet. |
| `partial` | Stats imported but not yet confirmed complete. |
| `final` | All player stats locked — no further changes expected. |

### Match result

One of: `home_win`, `away_win`, `draw`, `no_result`. Derived from club match scores; stored in `drv_result` once `data_status = final`.

### PlayerMatch status

| Status | Meaning                                                                                                                                                       |
|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `named` | Selected in the AFL team sheet. Match has not been played yet.                                                                                                |
| `played` | Played in the AFL match.                                                                                                                                      |
| `dnp` | Did not play — was in the squad but did not take the field. This can be a logical status, inferred from the fact that no Player Match exists for given Match. |

### Player tenure

`PlayerSeason` includes `from_round_id` and `to_round_id` to track when a player joined or left a club during the season (trades, delistings). Null means start/end of season respectively.

### Events published

- **`AFL.PlayerMatchUpdated`** — fired when a player's match stats change. Payload carries full stats (kicks, handballs, marks, hitouts, tackles, goals, behinds).
- **`AFL.MatchFinalized`** — fired when `afl.match.data_status` transitions to `final`. Triggers AFL match result derivation and AFL ladder recalculation. (Also signals the FFL service to recalculate affected club match scores.)

---

## FFL Context

Fantasy Football League — a fantasy competition built on AFL statistics.

### Positions (fantasy)

A **position** is a scoring slot in a fantasy team. It determines *which* AFL stat earns fantasy points and at what rate. Positions are **not** field positions (forward, midfielder, etc.).

| Position | Scores from | Multiplier | Starter slots |
|----------|-------------|------------|---------------|
| `goals` | Goals | 5 per goal | 3 |
| `kicks` | Kicks | 1 per kick | 4 |
| `handballs` | Handballs | 1 per handball | 4 |
| `marks` | Marks | 2 per mark | 2 |
| `tackles` | Tackles | 4 per tackle | 2 |
| `hitouts` | Hitouts | 1 per hitout | 2 |
| `star` | Goals + kicks + handballs + marks + tackles | 5×G + 1×K + 1×H + 2×M + 4×T | 1 |
| **Total** | | | **18** |

`PlayerMatch.CalculateScore(aflStats)` is a pure domain function that applies the position multiplier to AFL statistics.

### Team composition

A fantasy club submits a team each round. Teams need not be full.

#### Bench (up to 4 players)

| Bench role | Backup positions | Limit |
|------------|-----------------|-------|
| **Backup star** | `"star"` | at most 1 |
| **Dual-position** | exactly 2 non-star positions | at most 3 |

Hard rules enforced by `domain.ValidateTeam()`:
1. Starter count per position ≤ `PositionSlots[pos]`
2. Total bench players ≤ 4
3. At most 1 backup star (`BackupPositions` contains `"star"`)
4. Non-star bench players have **exactly 2** backup positions, neither of which is `"star"`
5. Each non-star position appears in **at most one** bench player's backup pair
6. At most 1 `InterchangePosition` set across all players in the team
7. `InterchangePosition` (if set) must be a recognised `Position` value

#### Bench player identification

| Role | How to identify |
|------|----------------|
| **Starter** | `BackupPositions == nil` |
| **Bench** | `BackupPositions != nil` |
| **Interchange** (bench subtype) | `BackupPositions != nil && InterchangePosition != nil` |

- **BackupPositions** — comma-separated list of positions this bench player can substitute into (`"star"` for bench star; two non-star positions for dual-position bench).
- **InterchangePosition** — the position this bench player comes on in when interchanging. Must be one of their own `BackupPositions`. A player cannot have `InterchangePosition` set without `BackupPositions`.

### FFL Player and AFL linkage

Every FFL player links to an AFL player. `Player.afl_player_id`, `PlayerSeason.afl_player_season_id`, and `PlayerMatch.afl_player_match_id` store the corresponding AFL row IDs. These are plain integers, not foreign keys (no cross-schema joins).

### ClubMatch data status

Tracks FFL teams. Key input for scoring calculations.

| Status | Meaning                                                                      |
|--------|------------------------------------------------------------------------------|
| `no_data` | Team not yet submitted for this round.                                       |
| `submitted` | Team imported. Player substitutions may still be pending resolution.         |
| `final` | Team confirmed after all subs resolved. Locked — no further changes expected. |

### Data status -> Score tiers

Combining AFL Match data status and FFL ClubMatch data status determines what can be calculated:

| AFL status | FFL status | Score tier |
|-----------|-----------|-----------|
| `partial` or `final` | `submitted` or `final` | **Provisional** — may change as stats arrive or manager resolves subs (which can alter team structure, including the interchange slot) |
| `final` | `final` | **Final** — locked; updates the official ladder |

### PlayerMatch status

FFL `PlayerMatch.status` is **not derived** — it may be initialised from AFL status but takes its own values.

| Status | Meaning |
|--------|---------|
| `named` | Selected in the AFL team sheet. Match has not been played yet. |
| `played` | Played in the AFL match. |
| `dnp` | Did not play — was in the squad but did not take the field. |

### Substitution and interchange

`ClubMatch.Score()` aggregates fantasy scores with two replacement rules, applied per starter slot:

1. **Substitution** — if a starter's status is `dnp`, a bench player whose `BackupPositions` includes that starter's position fills that slot. A player who played but earned 0 points **cannot** be substituted. A bench player may cover multiple positions but is consumed by at most one substitution.
2. **Interchange** — if a bench player's `InterchangePosition` matches a starter's position *and* the bench player's score strictly exceeds the starter's, they swap. Applies per individual starter slot within the position group.

Constraints:
- A bench player can only be used **once** (sub or interchange, not both).
- Substitution is evaluated before interchange.
- Interchange requires the bench player to **strictly outscore** the starter (ties keep the starter).
- Where multiple players are eligible for substitution into a DNP slot, the team owner chooses which bench player fills which position.

### Match style

FFL matches have a `match_style`:

| Style | Meaning |
|-------|---------|
| `versus` | Normal head-to-head match between two clubs. `home_club_match_id` and `away_club_match_id` are both set. |
| `bye` | Club does not play this round. Only `home_club_match_id` is set. |
| `super_bye` | All clubs still field a team, but premiership points are awarded by a yet-to-be-defined process. Uses `clubs` JSONB. |

For non-versus styles, `clubs` stores club_season_ids: `{"A": {"club_season_id": 2}, "B": {"club_season_id": 1}}`. `home_club_match_id` / `away_club_match_id` are nullable.

### Ladder

Same structure as AFL (played, won, lost, drawn, for, against, premiership points) with an additional **extra points** field for bonus/penalty adjustments.

### Events

**Subscribes:**
- `AFL.PlayerMatchUpdated` → incremental provisional score update for the affected player and club match.
- `AFL.MatchFinalized` → recalculate provisional/final FFL scores for all club matches in the round.

**Publishes:**
- `FFL.FantasyScoreCalculated` — carries the calculated score and the AFL PlayerMatch ID it was derived from.
- `FFL.TeamSubmitted` — fired when `ffl.club_match.data_status → submitted`. Triggers provisional score calculation.
- `FFL.TeamFinalized` — fired when `ffl.club_match.data_status → final`. Triggers final score calculation if AFL is also final.
- `FFL.ClubMatchScoreFinalized` — fired when a single club's score is locked (AFL final + FFL team final). Triggers check for full match finalization.
- `FFL.MatchFinalized` — fired when both clubs in an FFL match have finalized. Triggers `ffl.match.drv_result` derivation and ladder recalculation. Symmetric with `AFL.MatchFinalized`.

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

### Event Flow

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

Provisional ladder is computed on-demand from current scores — no dedicated event or separate columns.