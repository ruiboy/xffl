# Current Sprint — Phase 20: Data Management — Import Infrastructure

**Sprint goal:** Close out Phase 20 — schema health side quests, score reconciliation, and clean-up.
Season setup and historical import have moved to Phase 23.

---

## Side quest — Replace circular match↔club_match FKs with a role column

**Problem**: `afl.match.home_club_match_id` and `afl.match.away_club_match_id` create a circular FK
with `afl.club_match.match_id` — the parent row must be inserted before its children can exist, but
the children must exist before the FKs can be set. This forces an awkward insert-then-UPDATE pattern
and couples the match table to a specific two-team shape. Identical issue in `ffl.match` /
`ffl.club_match`, where the shape is even less fixed (e.g. superbye: multiple clubs, one match).

**Proposed approach**: Drop `home_club_match_id` / `away_club_match_id` from both `afl.match` and
`ffl.match`. Add a `role` column to `afl.club_match` and `ffl.club_match` (e.g. `home`, `away`; FFL
may later add `superbye` or other variants). All queries that currently navigate via the FK pair
instead filter/join on `role`.

- [x] Migration: drop FK columns from `afl.match`, add `role` to `afl.club_match`
- [x] Migration: drop FK columns from `ffl.match`, add `role` to `ffl.club_match`
- [x] Update AFL domain model, repository, and any queries that use home/away FK navigation
- [x] Update FFL domain model, repository, and any queries that use home/away FK navigation
- [x] Update seed data and integration tests

---

## Side quest — Enforce AFL FK integrity in FFL *(NOT NULL constraints + domain guards)*

**Problem**: FFL entities that reference AFL counterparts have no enforcement — a `ffl.round` with a
null `afl_round_id`, for example, is silently valid at the DB and domain layers. The one legitimate
exception is `ffl.player_match.afl_player_match_id`, which is intentionally nullable (a player can
be named without having played).

**Required NOT NULL columns** (all others in this group):
- `ffl.season.afl_season_id`
- `ffl.round.afl_round_id`
- `ffl.player.afl_player_id`
- `ffl.player_season.afl_player_season_id`

**Nullable by design**: `ffl.player_match.afl_player_match_id`

- [x] Migration: add NOT NULL constraint to the four columns above
- [x] Domain: enforce at construction time via domain invariants (return error / panic on nil afl ID for the four entities)
- [x] Verify existing seed data satisfies the constraints before migrating

---

## Side quest — Derived player match status *(fix state-transition bugs)*

**Problem**: `status` on both `afl.player_match` and `ffl.player_match` is set imperatively from
multiple scattered call sites (event handlers, recalc commands, import flows), each with slightly
different guards. Current symptom: FFL `dnp` incorrectly set when AFL match is not yet finalized —
`inferPlayerMatchStatuses` sees `afl_player_match_id = null` without knowing AFL match finality.

**Agreed design**:

Two entirely separate status concepts — conflating them was the root cause:

**AFL participation status** (`drv_afl_status` on `ffl.player_match`):
- Answers: "did this AFL player play in their AFL match?"
- Values: `playing` (stats exist, match not final) / `played` (stats exist, match final) / `dnp` (no stats, match final) / `null` (no import yet)
- AFL cannot emit `dnp` (no team-selection tracking); FFL infers it from absence after `AFL.MatchFinalized`
- Denormalised into `ffl.player_match.drv_afl_status`; single AFL domain function `ComputeAFLPlayerMatchStatus(matchDataStatus) string` returns `playing`/`played`

**FFL team position status** (`ffl.player_match.status`):
- Answers: "what is this player's role in the FFL team this round?"
- Values: `named` (default — in team, no TM override) / `subbed` / `interchanged`
- Set only by TM decisions; never touched by AFL import or score recalc
- `subbed`/`interchanged` deferred to the substitution/interchange sprint

**Stored ground truths**: only `afl.match.data_status` and `ffl.club_match.data_status`.
Everything else is derived via a single domain function per concern.

**Tasks**:

*AFL service*
- [x] Drop `afl.player_match.status` column — row existence is the played assertion; no status needed
- [x] Add `ComputeAFLPlayerMatchStatus(matchDataStatus MatchDataStatus) string` to AFL domain (`playing`/`played`)
- [x] Populate `PlayerMatchStats.Status` in Twirp response using that function
- [x] Remove `SetStatusForMatchID` call from `MarkMatchStatsFinal` (bulk `named→played` update disappears)
- [x] Unit-test `ComputeAFLPlayerMatchStatus` with a status table

*FFL service — schema*
- [x] Add `ffl.player_match.drv_afl_status` column (nullable; `playing`/`played`/`dnp`)
- [x] Migrate existing `ffl.player_match.status` values: `played`→`drv_afl_status=played`, `dnp`→`drv_afl_status=dnp`, `named`→`null`
- [x] Change `ffl.player_match.status` enum to `named`/`subbed`/`interchanged`; set all rows to `named`

*FFL service — domain and application*
- [x] Redefine `PlayerMatchStatus` type as `named`/`subbed`/`interchanged`; add `DrvAFLStatus` type
- [x] Update scoring logic — `ClubMatch.Score()` substitution eligibility checks `drv_afl_status = dnp`, not `status`
- [x] `ProcessPlayerMatchUpdated`: store AFL-computed status in `drv_afl_status` (drop direct `status` write)
- [x] `ProcessAFLRoundFinalized`: set `drv_afl_status = dnp` for all FFL players with `drv_afl_status IS NULL` in the round — replaces `inferPlayerMatchStatuses`
- [x] `RecalculateClubMatchScore`: update `drv_afl_status` for linked players only (playing/played from AFL stats); never touch unlinked players' status
- [x] Remove `inferPlayerMatchStatuses` entirely

**Architecture docs to update after this sprint** *(do not update now — ai/ is read-only during impl)*:
- `domain.md`: AFL PlayerMatch status section (remove `named`, document `playing`/`played`); FFL PlayerMatch status section (replace named/played/dnp table with new `status` + `drv_afl_status` tables); Substitution section (change DNP check to reference `drv_afl_status`)
- `event-flow.md`: `AFL.PlayerMatchUpdated` payload note (carries computed AFL status `playing`/`played`); FFL subscriber description (receives AFL status → writes `drv_afl_status`)

**Files to look at when starting**:
- `services/afl/internal/domain/player_match.go` — drop status, add `ComputeAFLPlayerMatchStatus`
- `services/afl/internal/application/data_ops.go` — remove `SetStatusForMatchID` from `MarkMatchStatsFinal`; Twirp layer to populate computed status
- `services/ffl/internal/domain/player_match.go` — redefine types, update `ClubMatch.Score()` substitution check
- `services/ffl/internal/application/score_commands.go` — replace `inferPlayerMatchStatuses` in `ProcessAFLRoundFinalized`
- `services/ffl/internal/application/commands.go` — `ProcessPlayerMatchUpdated`, `RecalculateClubMatchScore`

---

## Side quest — Team Manager substitution and interchange decisions

**Depends on**: Derived player match status (DNP must be reliable before choices are meaningful)

### Status model (agreed design)

`ffl.player_match.status` describes what happened to a player's **original named role**:

| Value | Applies to | Meaning |
|-------|-----------|---------|
| `named` | starters + bench | Role unchanged — starter played normally, or bench player (whether called upon or not) |
| `subbed` | starters only | Starter subbed out due to DNP; slot filled by bench player via `BackupPositions` |
| `interchanged` | starters only | Starter displaced by the interchange bench player via `InterchangePosition` |

Bench players are always `named`. Rule 5 guarantees at most one eligible bench player per
non-star position, so the sub/interchange pairing is always unambiguous from the starter's status
alone — no FK or extra field needed on the bench player.

### Scoring modes

`ClubMatch.Score()` detects two modes:

- **Auto mode** (all starters `named`): current behaviour — substitutes all DNP starters with
  first eligible bench player; applies interchange if bench player score exceeds starter's.
- **TM mode** (any starter has `subbed` or `interchanged`): status-driven —
  - `subbed` starter → replaced by the bench player whose `BackupPositions` covers their position
  - `interchanged` starter → replaced by the bench player whose `InterchangePosition` matches their position
  - `named` starter → scores normally

### DeclareSubs use case

Input: `clubMatchID`, `subbedOutIDs []int` (starter player_match IDs), `interchangeApplied bool`

Behaviour:
1. For each starter: set `status = subbed` if in `subbedOutIDs`, else reset to `named`
2. If `interchangeApplied`: find the bench player with `InterchangePosition` set; auto-determine
   which starter at that position to mark `interchanged` (lowest scorer, same logic as auto mode);
   set that starter to `interchanged`; reset any previously-interchanged starters
3. Trigger `RecalculateClubMatchScore`

Validation: each `subbedOutID` must be a starter with `aflStatus = dnp`.

### New mutation

```graphql
declareFFLSubstitutions(input: DeclareFFLSubstitutionsInput!): [FFLPlayerMatch!]!

input DeclareFFLSubstitutionsInput {
  clubMatchId: ID!
  subbedOutPlayerMatchIds: [ID!]!
  interchangeApplied: Boolean!
}
```

### UI — "Subs" mode in Team Builder

A third mode in the Team Builder (alongside read-only and Manage), active when the AFL match has
started (any player has `aflStatus` set). Shows:

- Starters grouped by position; DNP starters highlighted
- Each DNP starter with an eligible bench player gets a checkbox, pre-ticked by default
- Interchange toggle: shown if a bench player has `InterchangePosition` set; defaults to checked
  if that bench player's current score exceeds the relevant starter's score
- "Save Subs" button → calls `declareFFLSubstitutions`; decisions can be revised any time while
  club match status is `submitted`

### Tasks

*Domain*
- [x] Update `ClubMatch.Score()` for TM mode
- [x] Unit-test `Score()` TM mode (subbed, interchanged, mixed, edge cases)

*Application*
- [x] `DeclareSubs` use case in `Commands`
- [x] Integration-test `DeclareSubs`

*GraphQL*
- [x] `declareFFLSubstitutions` mutation + resolver

*Frontend*
- [x] "Subs" mode in Team Builder — DNP starter checkboxes, interchange toggle, Save Subs button

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
