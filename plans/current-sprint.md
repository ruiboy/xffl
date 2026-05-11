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

Once player DNP status is confidently derived, a Team Manager must be able to declare which bench
players cover which DNP starters (substitution) and whether any interchange swaps apply, within the
rules in `ai/architecture/domain.md`. This is a combined domain + UX concern — the order of
application is at the TM's discretion within the constraints. Detailed design deferred.

- [ ] Design and implement substitution/interchange decision model (domain + UI)

---

## Side quest — Pluggable FFL scoring formula *(prerequisite for Phase 23 historical backfill)*

- Different seasons use different scoring formulas (e.g. goals were worth 4 pts in some years, now different)
- Strategy pattern: implementations in code keyed by a string; each `ffl.season` maps to a strategy key
- Each strategy should carry a human-readable description (for frontend display)
- [ ] Design known formula variants and year ranges
- [ ] `ScoringStrategy` interface + concrete implementations
- [ ] `ffl.season.scoring_strategy` column (string key)
- [ ] Wire into score calculation use case

## Step 6 — Score reconciliation *(every round)*

**Rules:**
- UI: Team Manager gets to choose how to apply any player subsitutions from the bench, within the league rules.
- Submitted score = what the forum post recorded; `drv_score` = calculated from AFL stats
- Current season: `drv_score` is authoritative; generate a copy-pasteable forum summary of differences
- Previous seasons: submitted score is authoritative; record delta in `notes` column
- [ ] Add `notes TEXT` column to `ffl.club_match` and `ffl.player_match`
- [ ] `ReconcileScores` use case — compare submitted vs `drv_score`; produce structured diff
- [ ] FFL frontend — submitted vs calculated scores side by side; copy-pasteable forum summary output

## Close out

- [ ] Audit and remove `ffl.player.drv_name` — drop column from schema, domain, resolvers, frontend
- [ ] Retire `parse_forum.py`
- [ ] Move `afl.match.stats_import_status` + `stats_imported_at` out of core domain into `afl.dataops_match_source`
- [ ] Share `dev/postgres/test-e2e` init files with `dev/postgres/init` rather than duplicating
