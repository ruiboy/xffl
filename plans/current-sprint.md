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

**Architecture docs updated** *(done — see event-flow.md, domain.md, deferred.md)*

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

## Side quest — Event flow redesign *(fix bugs + clean architecture)*

**Goal:** Replace the current fragile event model with the design agreed in planning (see `ai/architecture/event-flow.md`). Fixes Bug 1 (playing→played never transitions), Bug 2 (premature DNP), and the premature `FFL.ClubMatchScoreFinalized` emission. Establishes clean, ordering-independent event semantics.

**Reference:** `ai/architecture/event-flow.md` — read before starting any task here.

---

### Contracts (`contracts/events/`)

- [x] Remove `AFL.MatchFinalized` constant and `AflMatchFinalizedPayload`
- [x] Remove `Status` field from `PlayerMatchUpdatedPayload`
- [x] Add `AFL.MatchUpdated` constant and `AflMatchUpdatedPayload` (`match_id`, `round_id`, `season_id`, `match_status`, `PlayerSeasonIDStatusMap map[int]string`)
- [x] Remove `FFL.TeamSubmitted` constant and `FflTeamSubmittedPayload`
- [x] Remove `FFL.TeamFinalized` constant and `FflTeamFinalizedPayload`
- [x] Add `FFL.ClubMatchUpdated` constant and `FflClubMatchUpdatedPayload` (`club_match_id`, `match_id`, `round_id`, `data_status`, `PlayerMatches map[int]FflPlayerMatchInfo`)
- [x] Add `FflPlayerMatchInfo` struct (`position`, `status`, `backup_positions`, `interchange_position`)
- [x] Rename `FFL.FantasyScoreCalculated` → `FFL.PlayerMatchUpdated`; rename `FantasyScoreCalculatedPayload` → `FflPlayerMatchUpdatedPayload`; add `club_match_id` field
- [x] Rename `FFL.MatchFinalized` → `FFL.MatchScoreFinalized`; rename `FflMatchFinalizedPayload` → `FflMatchScoreFinalizedPayload`
- [x] `FFL.ClubMatchScoreFinalized` and `FflClubMatchScoreFinalizedPayload` — no change

---

### AFL service

**`services/afl/internal/application/dataops.go`**
- [x] `ImportAFLStats`: remove `Status` from all `PlayerMatchUpdatedPayload` emissions
- [x] `ImportAFLStats`: after all player_matches written and `data_status → partial`, emit one `AFL.MatchUpdated(partial)` with `PlayerSeasonIDStatusMap` = all imported `afl_player_season_id → "playing"`
- [x] `MarkMatchStatsFinal`: remove `AFL.MatchFinalized` publish
- [x] `MarkMatchStatsFinal`: emit `AFL.MatchUpdated(final)` — build map: player_match rows → `"played"`, all player_seasons for both club_seasons not in player_match → `"dnp"`; inline match result derivation + ladder recalculation (no self-subscription needed)
- [x] `MarkMatchStatsFinal`: keep direct calls to derive match result + `RecalculateAFLLadder` (no change — these stay internal)

**`services/afl/internal/application/player_match.go`** (`UpdatePlayerMatch`)
- [x] Remove `Status` from `PlayerMatchUpdatedPayload`

**`services/afl/internal/interface/events/handlers.go`**
- [x] Remove `AFL.MatchFinalized` subscription (AFL no longer self-subscribes; finalization is handled directly)

---

### FFL service — domain

**`services/ffl/internal/domain/player_match.go`**
- [x] Remove `AFLStatusNamed` constant (pre-match named tracking is deferred — see `plans/roadmap.md`)

---

### FFL service — application

**`services/ffl/internal/application/score.go`** (or `queries.go` — wherever appropriate)
- [x] Add `AllAFLStatusesFinal(ctx, clubMatchID) bool` — returns true when every player_match in the club_match has `drv_afl_status ∈ {played, dnp}`; backed by a repository query

**`services/ffl/internal/application/reactions.go`**
- [x] Remove `ProcessAFLMatchFinalized` (replaced by `ProcessAFLMatchUpdated`)
- [x] Add `ProcessAFLMatchUpdated(ctx, AflMatchUpdatedPayload)`: applies status map via `applyAFLStatusMap`, recalculates score, emits `FFL.ClubMatchScoreFinalized` when both axes final
- [x] Remove `ProcessFflTeamFinalized` (replaced by `ProcessFflClubMatchUpdated`)
- [x] Add `ProcessFflClubMatchUpdated(ctx, ...)`: recalculates score, emits `FFL.ClubMatchScoreFinalized` when both axes final
- [x] Update `ProcessPlayerMatchUpdated` — no status field in payload; just link + score recalc + ladder cascade if both axes final
- [x] Rename `ProcessFflMatchFinalized` → `ProcessFflMatchScoreFinalized`
- [x] Update `emitClubMatchScoreFinalized` — no logic change (still emits `FFL.ClubMatchScoreFinalized`)
- [x] Note: `UpdateAFLStatusFromMap` implemented at application layer (`applyAFLStatusMap`) using existing `FindByIDs` + `UpdateAFLStatus` per player (sqlc unnest limitation)

**`services/ffl/internal/application/team.go`** (or `dataops.go`)
- [x] Replace `FFL.TeamSubmitted` publish → `FFL.ClubMatchUpdated(submitted)` with full player_matches snapshot
- [x] Replace `FFL.TeamFinalized` publish → `FFL.ClubMatchUpdated(final)` with full player_matches snapshot

**`services/ffl/internal/application/team.go`** (`DeclareSubs`)
- [x] After subs are applied: emit `FFL.ClubMatchUpdated(submitted)` with updated player_matches snapshot
- [x] Also emit `FFL.SubsDeclared` (new — no current subscriber; published for future consumers)

**`services/ffl/internal/application/score.go`**
- [x] `RecalculateScore`: removed AFL status sync (status comes only from `AFL.MatchUpdated` now)
- [x] `ProcessPlayerMatchUpdated` ladder cascade: after score recalc, if `AllAFLStatusesFinal` AND `data_status = final` → call `RecalculateFflLadder` directly

---

### FFL service — infrastructure

**`services/ffl/internal/infrastructure/postgres/sqlc/player_match.sql`**
- [x] Add `AllAFLStatusesFinal` query — returns bool: all player_matches for a club_match have `drv_afl_status IN ('played', 'dnp')`
- [x] Remove `SetDrvAFLStatusDNPForClubMatch` query
- [x] Regenerate sqlcgen after query changes (`sqlc generate`)
- [x] Note: `UpdateAFLStatusFromMap` implemented at application layer (sqlc parallel unnest not supported)

---

### FFL service — event handlers

**`services/ffl/internal/interface/events/handlers.go`**
- [x] Remove subscription to `AFL.MatchFinalized`
- [x] Add subscription to `AFL.MatchUpdated` → `ProcessAFLMatchUpdated`
- [x] Remove subscription to `FFL.TeamSubmitted`
- [x] Remove subscription to `FFL.TeamFinalized`
- [x] Add subscription to `FFL.ClubMatchUpdated` → `ProcessFflClubMatchUpdated`
- [x] Update `FFL.MatchFinalized` subscription → `FFL.MatchScoreFinalized` → `ProcessFflMatchScoreFinalized`
- [x] Update `FFL.FantasyScoreCalculated` subscription → `FFL.PlayerMatchUpdated` (Search handler)

---

### Search service

**`services/search/...`** *(wherever the event subscription lives)*
- [x] Update subscription from `FFL.FantasyScoreCalculated` → `FFL.PlayerMatchUpdated`

---

### Tests
- [x] Unit-test `AllAFLStatusesFinal` — all played, all dnp, mixed, null present, playing present
- [x] Integration-test `ProcessAFLMatchUpdated` (partial): verify `drv_afl_status = playing` set for correct players only; players in other AFL matches unaffected
- [x] Integration-test `ProcessAFLMatchUpdated` (final): verify played + dnp set correctly; `FFL.ClubMatchScoreFinalized` emitted when FFL team is also final
- [x] Integration-test `ProcessFflClubMatchUpdated` (final): verify `FFL.ClubMatchScoreFinalized` NOT emitted when AFL not yet final; emitted when AFL is final
- [x] Regression-test full event chain: import partial → import final → finalize FFL team → verify ladder

---

## Side quest — Explicit substitution/interchange records *(ffl.substitution_event)*

**Problem**: The current model infers sub/interchange pairings at query time from `backup_positions`
and `interchange_position` — the same heuristic logic is duplicated in `ClubMatch.Score()`,
`RecalculateScore`, the GraphQL resolvers, and the frontend. This makes scoring non-deterministic
(re-running the heuristic after squad changes can produce different pairings), prevents reliable
stats queries ("who covered whom across the season?"), and forces the frontend to reimplement
matching logic rather than reading a direct fact.

**Proposed approach**: Introduce `ffl.substitution_event` as the source of truth for every TM
sub/interchange decision. `DeclareSubs` writes these rows; `ClubMatch.Score()` and the frontend
read them directly.

### Schema

```sql
CREATE TABLE ffl.substitution_event (
  id                      serial PRIMARY KEY,
  club_match_id           integer NOT NULL REFERENCES ffl.club_match(id),
  type                    text    NOT NULL,  -- 'sub' | 'interchange'
  position                text    NOT NULL,  -- position key being filled
  replaced_pm_id          integer NOT NULL REFERENCES ffl.player_match(id),
  replacing_pm_id         integer NOT NULL REFERENCES ffl.player_match(id),
  created_at              timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ON ffl.substitution_event (club_match_id);
```

One row per pairing. Re-declaring subs deletes all existing rows for the `club_match_id` and
re-inserts. No versioning needed — only the latest declaration matters.

### Backend changes

**`DeclareSubs`**:
1. Delete all `substitution_event` rows for `club_match_id`.
2. For each `subbedOutID`: find the covering bench player (same pairing logic as today, but run
   once here), insert a `sub` row.
3. If `interchangeApplied`: find the displaced starter + interchange bench player, insert one
   `interchange` row.
4. Trigger `RecalculateClubMatchScore` as today.

**`ClubMatch.Score()` — TM mode**:
- Instead of re-running `backup_positions`/`interchange_position` heuristics, receive a
  `[]SubstitutionEvent` slice and build the replacement map from it.
- Auto mode (no substitution_event rows): behaviour unchanged.

**`RecalculateClubMatchScore`**:
- Load `substitution_event` rows for the club_match; pass to `ClubMatch.Score()`.

### GraphQL

```graphql
type FFLSubstitutionEvent {
  id:            ID!
  type:          String!   # "sub" | "interchange"
  position:      String!
  replacedBy:    FFLPlayerMatch!
  replacing:     FFLPlayerMatch!
}

extend type FFLClubMatch {
  substitutionEvents: [FFLSubstitutionEvent!]!
}
```

Frontend uses `substitutionEvents` instead of inferring pairings.

### Frontend simplification

- Drop computed refs: `subsMapping`, `savedSubsMap`, `interchangeDisplacedStarterNormal`,
  `savedSubsStarterMap`, `coveringMap` (in both `TeamBuilderView` and `SquadTable`).
- Replace with a map derived directly from `clubMatch.substitutionEvents`.
- `effectiveCovering` / `effectiveSubbedForStarter` become trivial lookups into that map.
- Subs mode (live, before Save): keep the client-side preview maps for immediate feedback;
  on save the server response includes updated `substitutionEvents` which drives normal mode.

### Tasks

*Migration*
- [ ] Write migration: create `ffl.substitution_event` table + index
- [ ] Write sqlc queries: `InsertSubstitutionEvent`, `DeleteSubstitutionEventsByClubMatch`,
      `GetSubstitutionEventsByClubMatch`
- [ ] Regenerate sqlcgen

*Domain*
- [ ] Add `SubstitutionEvent` value type to `ffl` domain
- [ ] Update `ClubMatch.Score()` signature to accept `[]SubstitutionEvent`; remove heuristic
      pairing logic; derive replacement map from events in TM mode

*Application*
- [ ] Update `DeclareSubs`: delete+reinsert substitution_event rows; derive pairings once here
- [ ] Update `RecalculateClubMatchScore`: load substitution_event rows; pass to `ClubMatch.Score()`
- [ ] Unit-test updated `ClubMatch.Score()` with explicit event slices
- [ ] Integration-test `DeclareSubs`: verify substitution_event rows written; re-declare replaces rows

*GraphQL*
- [ ] Add `FFLSubstitutionEvent` type and `substitutionEvents` field on `FFLClubMatch`
- [ ] Add sqlc-backed resolver for `substitutionEvents`

*Frontend*
- [ ] Add `substitutionEvents` to `GET_FFL_MATCH` and `GET_FFL_ROUND` queries
- [ ] Replace heuristic computed refs in `TeamBuilderView` with map derived from `substitutionEvents`
- [ ] Replace `coveringMap` / `coveredStarterMap` in `SquadTable` with map derived from `substitutionEvents`
- [ ] Verify subs mode live preview still works (client-side preview maps remain for in-flight state)

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
