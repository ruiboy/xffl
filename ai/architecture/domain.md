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

### Player match status

| Status | Meaning |
|--------|---------|
| `named` | Selected in the AFL team sheet. Match has not been played yet. |
| `played` | Played in the AFL match. |
| `dnp` | Did not play — was in the squad but did not take the field. |

### Player tenure

`PlayerSeason` includes `from_round_id` and `to_round_id` to track when a player joined or left a club during the season (trades, delistings). Null means start/end of season respectively.

### Match result

One of: `home_win`, `away_win`, `draw`, `no_result`.

### Events published

- **`AFL.PlayerMatchUpdated`** — fired when a player's match stats change. Payload carries full stats (kicks, handballs, marks, hitouts, tackles, goals, behinds).

---

## FFL Context

Fantasy Football League — a fantasy competition built on AFL statistics.

### Positions (fantasy)

A **position** is a scoring slot in a fantasy lineup. It determines *which* AFL stat earns fantasy points and at what rate. Positions are **not** field positions (forward, midfielder, etc.).

| Position | Scores from | Multiplier |
|----------|-------------|------------|
| `goals` | Goals | 5 per goal |
| `kicks` | Kicks | 1 per kick |
| `handballs` | Handballs | 1 per handball |
| `marks` | Marks | 2 per mark |
| `tackles` | Tackles | 4 per tackle |
| `hitouts` | Hitouts | 1 per hitout |
| `star` | Multiple | 5×goals + 1×kicks + 1×handballs + 2×marks + 4×tackles |

`PlayerMatch.CalculateScore(aflStats)` is a pure domain function that applies the position multiplier to AFL statistics.

### Lineup: starters and bench

A fantasy club fields **starters** and **bench** players each round. This distinction is structural, not stored in a status field:

| Role | How to identify |
|------|----------------|
| **Starter** | Occupies a position slot. Has neither `BackupPositions` nor `InterchangePosition` set. |
| **Bench** | Has `BackupPositions` and/or `InterchangePosition` set. Sits out unless substitution or interchange applies. |

- **BackupPositions** — comma-separated list of positions a bench player can substitute into.
- **InterchangePosition** — the single position a bench player competes against for an interchange swap.

### FFL Player and AFL linkage

Every FFL player corresponds to an AFL player. `Player.afl_player_id`, `PlayerSeason.afl_player_season_id`, and `PlayerMatch.afl_player_match_id` store the corresponding AFL row IDs. These are plain integers, not foreign keys (no cross-schema joins). `Player.name` is derived from the AFL player (`drv_name` in DB).

### Status

FFL `PlayerMatch.status` is **not derived** — it may be initialised from AFL status but takes its own values.

| Status | Meaning |
|--------|---------|
| `named` | Selected in the AFL team sheet. Match has not been played yet. |
| `played` | Played in the AFL match. |
| `dnp` | Did not play — was in the squad but did not take the field. |

### Substitution and interchange

`ClubMatch.Score()` aggregates fantasy scores with two replacement rules:

1. **Substitution** — if a starter's status is `dnp`, a bench player whose `BackupPositions` includes that starter's position fills in. A bench player may cover multiple positions but only subs into one.
2. **Interchange** — if a bench player's `InterchangePosition` targets a starter *and* the bench player's score strictly exceeds the starter's, they swap.

Constraints:
- A bench player can only be used **once** (sub or interchange, not both).
- Interchange requires the bench player to **strictly outscore** the starter (ties stay).
- The **order** of substitution vs interchange, and which position a multi-position bench player fills, is decided by the team owner after AFL stats are in. The exact mechanism for this is TBD.

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

- **Subscribes:** `AFL.PlayerMatchUpdated` → triggers fantasy score calculation.
- **Publishes:** `FFL.FantasyScoreCalculated` — carries the calculated score and the AFL PlayerMatch ID it was derived from.