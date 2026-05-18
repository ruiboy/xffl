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

## Side quest — Explicit substitution/interchange records

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
Score = sum(pm.score) where pm.status IN ('named', 'subbed_in', 'interchanged_in')
        AND pm.backupPositions IS NULL   -- excludes named bench players
```

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
- [ ] Update `ai/architecture/domain.md` PlayerMatch status table with the five new values

*Domain*
- [ ] Update `PlayerMatchStatus` enum with the four new values
- [ ] Replace `scoreAuto()`, `scoreTM()`, `isTMMode()` with a single `Score()`:
      sum where `status IN (named, subbed_in, interchanged_in) AND backupPositions IS NULL`
- [ ] Update `DeclareSubs` domain method to accept explicit `[]SubPairing` + optional interchange
      pairing; validate and return updated `[]PlayerMatch` with new statuses set on both sides
- [ ] Unit-test `Score()` with all status combinations
- [ ] Unit-test `DeclareSubs` validation (invalid replacedPmId, wrong bench player, re-declare)

*Application*
- [ ] Update `DeclareSubs` application service: reset prior statuses, call domain method, persist
- [ ] Update `RecalculateClubMatchScore`: no changes needed (Score() is now self-contained)
- [ ] Integration-test `DeclareSubs`: verify player_match statuses set correctly; re-declare resets

*GraphQL*
- [ ] Update `declareSubs` mutation input to `FFLSubPairing` shape
- [ ] Update `playerMatch.status` docs/schema to reflect new values

*Frontend*
- [ ] Update `onSaveSubs` in `TeamBuilderView`: build `{replacedPmId, replacingPmId}` pairs from
      existing `subsMapping` ref; send to updated mutation
- [ ] Display mode in `TeamBuilderView`: replace `savedSubsMap`, `interchangeDisplacedStarterNormal`,
      `savedSubsStarterMap`, `effectiveCovering`, `effectiveSubbedForStarter` with direct `status` reads
- [ ] `SquadTable`: replace `coveringMap` / `coveredStarterMap` heuristics with direct `status` reads
- [ ] Update `initSubsState` to seed from `subbed_out` / `interchanged_out` statuses
- [ ] Update e2e tests: `declareSubs` mutation shape + any assertions on player status values

---

## Step 6 — Score reconciliation *(every round)*

*Requirements TBD — interview user when we reach this step before any implementation. Ideas to seed the conversation:*

- *What does "submitted score" mean vs the calculated `drv_score`? Who owns each?*
- *Current season: should calculated score be authoritative, or does the TM get final say?*
- *Previous seasons: if submitted score is the record, how do we track the delta?*
- *Is a copy-pasteable forum summary useful, or is in-app diff enough?*
- *Do we need a `notes` field on `club_match` / `player_match` for manual overrides or commentary?*

## Follow-on — Real data & ladder *(after Phase 20 close-out)*

**Goal:** Replace synthetic seed data with a full current-season dataset, unlock ladder generation, and smoke-test the end-to-end data ops / team submission / substitution flow against real data.

### 6a — Backup / restore reliability
- [ ] Verify `just backup-db` and `just restore-db` round-trip cleanly (no data loss, no schema drift)
- [ ] Document any gaps; fix before pulling live data

### 6b — Pull current-season AFL & team data
- [ ] Import all 2025 AFL rounds played to date (teams, players, match stats) via existing import tooling
- [ ] Import all 2025 FFL round teams to date
- [ ] Verify ladder calculation produces correct standings
- [ ] Smoke-test team submission and substitution flows against real round data; capture any edge cases / bugs found

### 6c — Retire full seed; keep slim demo seed
- [ ] Reduce `dev/seed` to a single representative round (enough for demonstration and future dev work)
- [ ] Confirm e2e tests still pass against slim seed
- [ ] Live data becomes the source of truth; seed is development scaffold only

## Close out

- [ ] Audit and remove `ffl.player.drv_name` — drop column from schema, domain, resolvers, frontend
- [ ] Retire `parse_forum.py`
- [ ] Move `afl.match.stats_import_status` + `stats_imported_at` out of core domain into `afl.dataops_match_source`
- [ ] Share `dev/postgres/test-e2e` init files with `dev/postgres/init` rather than duplicating
