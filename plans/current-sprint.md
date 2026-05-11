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

- [ ] Migration: add NOT NULL constraint to the four columns above
- [ ] Domain: enforce at construction time via domain invariants (return error / panic on nil afl ID for the four entities)
- [ ] Verify existing seed data satisfies the constraints before migrating

---

## Side quest — Derived player match status *(fix state-transition bugs)*

**Problem**: `status` on both `afl.player_match` and `ffl.player_match` is set imperatively from
multiple scattered call sites (event handlers, recalc commands, import flows), each with slightly
different guards. Correctness depends on call order and which data has been populated. Current
symptoms: FFL DNP not being set when `ffl.match.home_club_match_id` is null (fixed); FFL DNP
incorrectly set when AFL match is not yet finalized; AFL `named→played` transition done
separately from FFL status inference.

**Key insight**: status is *fully determined* by two ground-truth values that are already in the DB:

```
afl.player_match:
  status = (did AFL player match row exist?) → played / dnp (inferred from absence within a finalized match)

ffl.player_match:
  afl_player_match_id IS NOT NULL              → played
  afl_player_match_id IS NULL
    AND AFL match data_status = 'final'        → dnp
    AND AFL match data_status ≠ 'final'        → named
```

**Proposed approach**:

1. **Single domain function** `ComputePlayerMatchStatus(aflPlayerMatchID *int, aflMatchDataStatus string) PlayerMatchStatus`
   encodes the derivation once. All imperative `UpdateStatus` call sites are replaced with a call
   to this function followed by a single persist.

2. **Thread AFL match data_status into every status-setting path** — the missing input that currently
   causes `inferPlayerMatchStatuses` to set DNP before a match has been played (it sees
   `afl_player_match_id = null` without knowing whether the AFL match is even in progress).

3. **One recalculation entry point** — `RecalculateClubMatchScore` re-derives status from scratch as
   part of its normal work. `inferPlayerMatchStatuses` as a separate pass is removed; status falls
   out naturally from the recalc.

4. **Same principle for AFL** — AFL `afl.player_match.status` (`named`→`played`) follows the same
   pattern: derive from whether a player match row exists within a finalized match, rather than
   setting it as a side effect of the import flow.

5. **Unit-test the derivation function** with a table of (aflPlayerMatchID, aflMatchDataStatus) →
   expected status cases. Integration tests verify data flows in correctly, not the status logic.

**Files to look at when starting**:
- `services/ffl/internal/application/score_commands.go` — `inferPlayerMatchStatuses` (to be replaced)
- `services/ffl/internal/application/commands.go` — `ProcessPlayerMatchUpdated`, `RecalculateClubMatchScore`
- `services/afl/internal/application/data_ops.go` — AFL status set during import
- `services/ffl/internal/domain/player_match.go` — `PlayerMatchStatus` type (home for the new function)

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
