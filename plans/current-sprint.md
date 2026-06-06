# Current Sprint — Phase 20: Data Management — Import Infrastructure

**Sprint goal:** Close out Phase 20 — schema health side quests, score reconciliation, and clean-up.
Season setup and historical import have moved to Phase 23.

---

## Side quest — Replace circular match↔club_match FKs with a role column - DONE

**Problem**: `afl.match.home_club_match_id` and `afl.match.away_club_match_id` create a circular FK
with `afl.club_match.match_id` — the parent row must be inserted before its children can exist, but
the children must exist before the FKs can be set.

---

## Side quest — Enforce AFL FK integrity in FFL - DONE

**Problem**: FFL entities that reference AFL counterparts have no enforcement — a `ffl.round` with a
null `afl_round_id`, for example, is silently valid at the DB and domain layers. The one legitimate
exception is `ffl.player_match.afl_player_match_id`, which is intentionally nullable (a player can
be named without having played).

---

## Side quest — Derived player match status *(fix state-transition bugs)* - DONE

**Problem**: `status` on both `afl.player_match` and `ffl.player_match` is set imperatively from
multiple scattered call sites (event handlers, recalc commands, import flows), each with slightly
different guards.
**Agreed design**:

Two entirely separate status concepts — conflating them was the root cause:

**AFL participation status** (`drv_afl_status` on `ffl.player_match`):
**FFL team position status** (`ffl.player_match.status`):

---

### UI — "Subs" mode in Team Builder - DONE

A third mode in the Team Builder (alongside read-only and Manage), active when the AFL match has
started (any player has `aflStatus` set).

---

## Side quest — Event flow redesign *(fix bugs + clean architecture)* - DONE

**Goal:** Replace the current fragile event model with the design agreed in planning (see `ai/architecture/event-flow.md`).

---

## Side quest — Explicit substitution/interchange records - DONE

**Problem**: The current model infers sub/interchange pairings at query time from `backup_positions`
and `interchange_position` — the same heuristic logic is duplicated in `ClubMatch.Score()`,
`RecalculateScore`, the GraphQL resolvers, and the frontend. This makes scoring non-deterministic
and forces the frontend to reimplement matching logic rather than reading a direct fact.

**Agreed design**: No heuristics, no auto mode, no separate event table. TM declares all
sub/interchange decisions explicitly via the UI. `DeclareSubs` mutates `player_match` rows
directly using new status values. `ClubMatch.Score()` becomes a trivial status read.
DNP players with no TM declaration score zero.

### Status values (replaces `named | subbed | interchanged`)

| Value | Who | Scores? |
|---|---|---|
| `named` | starter playing; unused bench | starters yes, bench no |
| `subbed_out` | starter explicitly replaced by TM | no |
| `subbed_in` | bench player brought in by TM | yes |
| `interchanged_out` | starter explicitly displaced by TM | no |
| `interchanged_in` | interchange bench player activated by TM | yes |

### Scoring — no modes, no heuristics

```
Score = sum(pm.score) where pm.position IS NOT NULL
        AND pm.status IN ('named', 'subbed_in', 'interchanged_in')
```

Inactive bench players always have `position = NULL`; activated bench players inherit the
starter's position (sub) or their declared interchange position (interchange). These two
conditions are orthogonal — no `backupPositions` check needed.

### `DeclareSubs` input (explicit pairings from TM via frontend)

```graphql
input FFLSubPairing {
  replacedPmId:  Int!   # starter going out
  replacingPmId: Int!   # bench player coming in
}

declareSubs(
  clubMatchId: Int!
  subs:        [FFLSubPairing!]!
  interchange: FFLSubPairing
): FFLClubMatch!
```

Re-declaration: reset all `subbed_out/subbed_in/interchanged_out/interchanged_in` rows for the
club match, then apply the new pairings.

### Frontend subs UX (unchanged from user's perspective)

The frontend derives pairings deterministically from `backupPositions` (one bench player per
position) and shows them as a preview. On Save, it sends the explicit `{replacedPmId,
replacingPmId}` pairs derived from that preview — no ambiguity, no picker UI needed.

Display mode: read `status` directly from each `playerMatch`; no heuristic inference needed.

### Tasks

*Migration*
- [ ] Add `subbed_out`, `subbed_in`, `interchanged_out`, `interchanged_in` to the status domain
      and migrate existing rows: `subbed → subbed_out`, `interchanged → interchanged_out`

*Docs*
- [x] Update `ai/architecture/domain.md` PlayerMatch status table with the five new values

*Domain*
- [x] Update `PlayerMatchStatus` enum with the four new values
- [x] Replace `scoreAuto()`, `scoreTM()`, `isTMMode()` with a single `Score()`:
      sum where `status IN (named, subbed_in, interchanged_in) AND backupPositions IS NULL`
- [x] Update `DeclareSubs` domain method to accept explicit `[]SubPairing` + optional interchange
      pairing; validate and return updated `[]PlayerMatch` with new statuses set on both sides
- [x] Unit-test `Score()` with all status combinations
- [x] Unit-test `DeclareSubs` validation (invalid replacedPmId, wrong bench player, re-declare)

*Application*
- [x] Update `DeclareSubs` application service: reset prior statuses, call domain method, persist
- [x] Update `RecalculateClubMatchScore`: no changes needed (Score() is now self-contained)
- [x] Integration-test `DeclareSubs`: verify player_match statuses set correctly; re-declare resets; bench player inherits position and scores correctly after RecalculateScore

*GraphQL*
- [x] Update `declareSubs` mutation input to `FFLSubPairing` shape
- [x] Update `playerMatch.status` docs/schema to reflect new values

*Frontend*
- [x] Update `onSaveSubs` in `TeamBuilderView`: build `{replacedPmId, replacingPmId}` pairs from
      existing `subsMapping` ref; send to updated mutation
- [x] Display mode in `TeamBuilderView`: replace `savedSubsMap`, `interchangeDisplacedStarterNormal`,
      `savedSubsStarterMap`, `effectiveCovering`, `effectiveSubbedForStarter` with direct `status` reads
- [x] `SquadTable`: replace `coveringMap` / `coveredStarterMap` heuristics with direct `status` reads
- [x] Update `initSubsState` to seed from `subbed_out` / `interchanged_out` statuses
- [x] Update e2e tests: `declareSubs` mutation shape + any assertions on player status values (none found)

---

## Side quest — AFL Byes

**Goal:** Support AFL bye rounds in both the AFL and FFL contexts — clubs on a bye have no match, but their eligible FFL players can still score via season average.

### Agreed design

**AFL context**

- A new **Bye** entity (`afl.bye`) records a club's absence from a round. It is a first-class entity, not a variant of Match — no `kind` column on `afl.match` is needed.
- Bye clubs have no `ClubMatch` or `PlayerMatch` records for that round. No AFL data is imported for bye rounds.

**FFL context — AFL Status**

- `drv_afl_status = "bye"` is added as a new AFL Status value on `ffl.player_match`. It is derived from AFL data (the club has a bye this round); never set by TM.
- TM status remains clean — the TM still sets the player as `named`. AFL Status = `"bye"` changes the *scoring input*, not the *gating condition* (position + TM status still determines whether a player contributes).

**Eligibility**

- A bye player may only be named if they played in their club's most recent non-bye AFL match.
- Validated at team submission; an ineligible player cannot be named.
- No eligibility flag needs to be stored — it is enforced at submission and queryable from AFL history.

**Scoring**

- Score is the player's season-to-date average for their position, with each AFL stat **floored per stat** before the multiplier is applied.
  - Non-star: `floor(avg_stat) × multiplier`
  - Star: `floor(avg_goals)×5 + floor(avg_kicks)×1 + floor(avg_handballs)×1 + floor(avg_marks)×2 + floor(avg_tackles)×4`
- For starters: `drv_score` is set at team submission (same field, sourced from average instead of match stats).
- For bench players: `drv_score` is set at sub declaration time, once the activated position is known. No new DB column needed.

### Tasks

*AFL domain*
- [x] Add `Bye` entity and repository (`FindByRoundID`, `FindByRoundAndClub`, `Upsert`)
- [x] Add `afl.bye` table to init script (no migration needed)
- [x] Seed `afl.bye` rows for all rounds — bye clubs are already noted in comments on each round block in `dev/postgres/seed/01_afl_seed.sql`; derive from clubs absent from that round's matches

*FFL domain*
- [x] Add `bye` to `AFLStatus` enum
- [x] No new DB column needed for bye scores: starters have `drv_score` set at submission (from season average); bench players have `drv_score` set at sub declaration time (compute from AFL history at the position they activate into)
- [x] Add `CalculateByeScore(position, avgStats)` pure domain function — unit tested
- [x] Update team submission validation: reject naming a bye player who didn't play last game
- [x] Update score calculation: when `drv_afl_status = "bye"`, `drv_score` is already set at submission (starter) or at sub declaration (bench) — no separate lookup needed

*Application*
- [x] On team submission for a bye round: for each bye-club starter, compute season-average stats, call `CalculateByeScore`, write to `drv_score`, set `drv_afl_status = "bye"`
- [x] On `DeclareSubs` for a bye round: for each activated bench player whose club has a bye, compute season-average stats at their assigned position, call `CalculateByeScore`, write to `drv_score`
- [x] Integration-test: bye player named + eligible → scores via average; ineligible → submission rejected; bench bye player activated → correct position score used

*GraphQL / frontend*
- [x] Add `bye` to `FFLAFLPlayerMatchStatus` enum in `query.graphqls` and regenerate — field, resolver, and converter are already wired
- [x] Team Builder: show bye indicator and season-average score for bye players

---

## Side quest — Consistent display ordering

**Goal:** Make row ordering deterministic and meaningful everywhere in the UX. Avoid `display_order` columns unless no natural key exists.

**Agreed approach per context:**

| Context | Ordering |
|---|---|
| AFL matches within a round | `start_dt ASC` — already a TIMESTAMP; ensure import populates times correctly |
| FFL matches within a round | Home club name `ASC` (stable, no real-world time exists) |
| Clubs within a match | `side`: home first, away second |
| AFL players in a match | Player name `ASC` |
| FFL `player_match` within a position group | `display_order ASC` — user-defined; only place a `display_order` column is needed |

**Tasks**
- [x] Apply `ORDER BY start_dt ASC` to AFL match queries; verify import populates match times (not just dates)
- [x] Apply home-first ordering to FFL match queries (ORDER BY id — frontend owns name-based sort)
- [x] Apply `side` ordering to club_match queries (AFL and FFL)
- [x] Apply name ordering to AFL player-in-match queries (ORDER BY id — frontend owns name sort)
- [x] Add `display_order` to `ffl.player_match`; wire through domain → repository → application → GraphQL → frontend (see separate player ordering side quest)

---

## Step 6 — Score reconciliation *(every round)* - DONE

*Agreed design (interviewed 2026-05-31):*

- `drv_score` is authoritative.
- `notes TEXT` added to `ffl.club_match` and `ffl.player_match` (before `drv_score`) as a free-form
  escape hatch for capturing manual override rationale, injuries, anomalies, etc.
- No `submitted_score` field for now — pattern unclear; promote to a proper column once the shape is known.
- No copy-pasteable forum summary needed.

## Follow-on — Real data & ladder *(after Phase 20 close-out)*

**Goal:** Replace synthetic seed data with a full current-season dataset, unlock ladder generation, and smoke-test the end-to-end data ops / team submission / substitution flow against real data.

### 6a — Backup / restore reliability
- [x] Verify `just backup-db` and `just restore-db` round-trip cleanly (no data loss, no schema drift)
- [x] Document any gaps; fix before pulling live data

### 6b — Pull current-season AFL & team data
- [ ] Import all 2026 AFL rounds played to date (teams, players, match stats) via existing import tooling
- [ ] Import all 2026 FFL round teams to date
- [ ] Verify ladder calculation produces correct standings
- [ ] Smoke-test team submission and substitution flows against real round data; capture any edge cases / bugs found

### 6c — Retire full seed; keep slim demo seed
- [ ] Reduce `dev/seed` to a single representative round (enough for demonstration and future dev work)
- [ ] Confirm e2e tests still pass against slim seed
- [ ] Live data becomes the source of truth; seed is development scaffold only

## Close out

- [x] Audit and remove `ffl.player.drv_name` — drop column from schema, domain, resolvers, frontend
- [x] Retire `parse_forum.py`
- [x] Move `afl.match.stats_import_status` + `stats_imported_at` out of core domain into `afl.dataops_match_source`
- [x] Share `dev/postgres/test-e2e` init files with `dev/postgres/init` rather than duplicating
