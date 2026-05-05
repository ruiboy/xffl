# Current Sprint — Phase 20: Data Management — Import Infrastructure

**Sprint goal:** Build recurring data flows for team submissions, AFL stats, score reconciliation, historical backfill, and season setup. All Go; ports-and-adapters throughout; Twirp for cross-service calls.

ADR: ADR-018 (Twirp for cross-service communication)

---

## Cross-cutting decisions

- All Go — no Python in production; single binary deployment
- `TeamParser`, `StatsParser`, `PlayerResolver` are application-layer interfaces; adapters live in infrastructure — input source never touches use case logic
- FFL service calls AFL service via Twirp to resolve `afl_player_id` and look up players; proto definitions in `contracts/`
- `PlayerResolver` uses club code to narrow candidates before fuzzy name matching; confidence threshold gates auto-commit vs. review queue
- Frontend import feature lives in `features/data-ops/`

## Data model decisions

- `ffl.player.drv_name` is retired in principle — all Phase 20 code must not read or write it; audit usages and remove (schema + domain + resolvers + frontend) when we reach close-out
- Player names are owned by the AFL service; the FFL service fetches them via a batch Twirp call when building a candidate pool for matching — no denormalisation
- `drv_` columns elsewhere in the schema (scores, ladder) are legitimate derived/computed values and are not affected by this decision
- `ffl.player_match.afl_player_match_id` may be null at submission time (AFL stats not yet available); linked later when stats are imported
- `ffl.match` is pre-created (fixture); team submission creates/updates `ffl.club_match` and `ffl.player_match` records against it
- `ffl.club_match` and `ffl.player_match` need a free-text `notes` column for reconciliation commentary (score deltas, manual overrides)

---

## Step 4 — Round team submission *(every round)*

- [x] ADR-018 — Twirp for cross-service communication
- [x] Twirp: proto + buf toolchain; batch `PlayerLookup` handler on AFL service; FFL `infrastructure/rpc/` adapter
- [x] `TeamParser` port interface (application layer)
- [x] `PlayerResolver` port interface (application layer)
- [x] `ParseTeamSubmission` use case
- [x] `ImportRoundTeams` use case
- [x] `ForumPostParser` adapter (infrastructure)
- [x] FFL GraphQL: `parseTeamSubmission` + `confirmTeamSubmission` mutations
- [x] Frontend: `features/data-ops/` — club + round dropdowns, paste form, review table, confirm button
- [x] Tests: parser unit tests + GraphQL integration test + e2e golden path

## Step 5 — AFL stats import *(automated)*

- [x] `StatsParser` port interface
- [x] `ImportAFLStats` use case
- [x] FootyWire scraper adapter
- [x] `MatchSourceMapRepository` ACL table (ADR-016)
- [x] `MarkMatchStatsComplete` use case + mutation
- [x] AFL GraphQL mutations: `importAFLMatchStats`, `markAFLMatchStatsComplete`

## Step 5a — Data ops management UI

- [x] Data ops UI: FFL rounds list with team/stats status indicators
- [x] AFL Stats Import tab with per-match scrape + mark complete

## Step 3 — In-season player trades *(frequent)*

**Decisions:**
- `ffl.season.afl_season_id` FK required — foundational for scoping player season lookup; schema migration and seed already complete
- **Remove:** sets `to_round_id` on `ffl.player_season` (preserves history); UI shows a round dropdown (returns explicit `round_id`); defaults to current round but allows past/future
- **Add:** graph-traversal backed; FFL season links to AFL season via `afl_season_id`; gateway queries AFL service for player seasons scoped to that AFL season; no Typesense involvement
- `ADD_FFL_SQUAD_PLAYER` mutation needs extending to accept `aflPlayerSeasonId` and `fromRoundId`; backing use case extended accordingly
- Entry point remains the existing Manage mode on SquadView

**Tasks:**
- [x] Schema migration: `ffl.season.afl_season_id INTEGER REFERENCES afl.season(id)`
- [x] Seed `dev/postgres/seed/02_ffl_seed.sql` with `afl_season_id` for FFL 2026 → AFL 2026
- [x] AFL service: `AFLSeason.playerSeasons(filter, first, after)` connection; `FFLSeason.aflSeason` traversal via federation
- [x] Extend `ADD_FFL_SQUAD_PLAYER` mutation + use case: accept `aflPlayerSeasonId`, `fromRoundId`
- [x] Extend remove mutation + use case: accept `toRoundId` instead of hard-delete
- [x] SquadView: remove button → round dropdown + confirm; add panel → graph-backed player season search
- [x] Make sure to/from round id is recorded for all player trades; tighten UI
- Side quest:
  - [x] Add ffl.player_season.notes and cost columns.
  - [x] Add a little player season dialog showing to/from round, cost, and allowing to edit notes in Manage and normal modes
- [x] Tighten graph endpoints for player trades: `addFFLPlayerToSeason`, `removeFFLPlayerFromSeason`, `addFFLSquadPlayer`
- [x] Update e2e tests for player trade flows (after above is done)
- [x] Retire `ffl.player.drv_name`: add `FFLPlayer.aflPlayer: AFLPlayer` federation traversal; audit all `player { name }` reads in frontend and switch to `player { aflPlayer { name } }`; deprecate `FFLPlayer.name`; drop `drv_name` column. Until done, new `ffl.player` rows have `drv_name=""` so squad/match views show empty names for newly-added players.
- Side quest:
  - [x] Streamline supergraph: file-based composition (no running services needed) + Apollo Router `--hot-reload`
- Side quest - Data Ops:
  - AFL stats import and FFL Team import pages to align more closely in intent and UX:
    - AFL Stats import
      - [x] ~~On import of stats, if AFL PLayer can not be found, do a name matching thing like FFL Team import.~~ → replaced by holistic player search (see below)
      - [ ] Link round to AFL Round page.
      - [x] Link match to AFL Match page.
      - [x] AFL match has stats_import_status / ts tracking columns; these could be genericised to Status = no data, partial stats, final stats. No timestamp. Not tied to "import" as such, but set by import (and maybe other things later).
      - [ ] Set status of all matches where stats are imported in data seed. 
    - FFL Team import
      - [x] Round page should list all FFL teams and show current status (similar to AFL Stats import page). Then each row facilitates import somehow.
      - [x] FFL team (= club match) should have status tracking, not tied to import, but set by import. Status = team submitted, portial score (?), final score.
      - [ ] Link round to FFL Round page.
      - [ ] Link each team to FFL Team Builder page
      - [x] Improve UI: Form is a bit ugly right now. Team format should default selected club.
    - [ ] Function to recalculate FFL stats for a team. Call on team import. Maybe have button in data ops.
  - Holistic player search (replaces inline unmatched-player review in AFL stats import)
    **Decisions:**
    - Player search is a shared UX pattern but two separate components: `features/data-ops/` and `features/ffl/` each own theirs so they can diverge naturally
    - Both components display latest AFL club name + season name alongside player name to aid disambiguation
    - Selecting a player from the list is the mapping action; "Add new player" button is used for when the player isn't found
    - `afl.dataops_player_source` table (mirrors `afl.dataops_match_source`): PK `(source, external_season, external_club, external_player)` → `player_season_id`; club included because footywire identity is per-club; row only written when a name mismatch was manually resolved — not for natural name matches; looked up on import before fuzzy matching so future imports auto-resolve
    - AFL Stats Import player flow: search → select existing OR add new → `resolveAFLPlayerMatch` mutation (creates `dataops_player_source` row if name differed + upserts `afl.player_match`)
    - FFL Add Player flow: same search → select existing (`addFFLPlayerToSeason`, existing) OR add new (two sequential frontend calls: `addAFLPlayer` on AFL service → `addFFLPlayerToSeason` on FFL service with returned `aflPlayerSeasonId`); FFL clubs have no correlation with AFL clubs so FFL service cannot derive AFL context
    - "Add new player" modal in FFL includes name input + AFL club season dropdown (user specifies AFL club explicitly)
    - `addAFLPlayer(input: AddAFLPlayerInput!)` takes `name` + `clubSeasonId` only — season is implied by club season; lives on AFL service
    - Remove from commit 5f8fcaf: `AFLPlayerCandidate` type, `candidates` field on `UnmatchedAFLPlayer`, confidence-ranked dropdown, inline expandable table in DataOpsView, `CONFIRM_AFL_PLAYER_MATCH` frontend mutation
    **Tasks:**
    - [x] `afl.dataops_player_source` schema + migration (AFL init + test-e2e init)
    - [x] AFL service: `PlayerSourceMapRepository` port + postgres adapter; lookup wired into import before fuzzy match
    - [x] AFL service: `AFLPlayer.latestPlayerSeason` field (resolver: highest season year for that player)
    - [x] AFL service: simplify `ImportAFLMatchStats` result — strip `candidates` from `UnmatchedAFLPlayer`, keep name + stats + clubMatchId
    - [x] AFL service: `addAFLPlayer(name, clubSeasonId)` use case + mutation → creates `afl.player` + `afl.player_season`
    - [x] AFL service: `resolveAFLPlayerMatch(clubMatchId, playerSeasonId, stats, sourceMapping?)` use case + mutation → writes `dataops_player_source` if mapping provided + upserts `afl.player_match`
    - [x] Frontend data-ops: remove inline candidate table; new `PlayerSearchModal.vue`; rework unmatched-player section (Resolve button per row)
    - [x] Frontend FFL: new `PlayerSearchModal.vue`; extract search logic from SquadView; "Add new player" calls `addAFLPlayer` then `addFFLPlayerToSeason` sequentially
  
## Step 6 — Score reconciliation *(every round)*

**Rules:**
- Submitted score = what the forum post recorded; `drv_score` = calculated from AFL stats
- Current season: `drv_score` is authoritative; generate a copy-pasteable forum summary of differences
- Previous seasons: submitted score is authoritative; record delta in `notes` column
- [ ] Add `notes TEXT` column to `ffl.club_match` and `ffl.player_match`
- [ ] `ReconcileScores` use case — compare submitted vs `drv_score`; produce structured diff
- [ ] FFL frontend — submitted vs calculated scores side by side; copy-pasteable forum summary output

## Step 2 — FFL squad import *(once/season)*

- Prerequisite: AFL season player records already exist (Step 1, or seeded from stats imports)
- [ ] `ImportFFLSquad` use case (FFL service) — match AFL player IDs from existing `afl.player` records; create `ffl.player` + `ffl.player_season` records
- [ ] `just import-ffl-squad` CLI trigger
- [ ] FFL frontend admin page — proposed player mappings for accept/reject

## Step 1 — AFL season player import *(once/season)*

- Source data TBD; UX must handle: matching to existing players (allowing for club changes), retired players, and brand new players
- [ ] Design import UX (accept/reject/skip flow for each proposed match)
- [ ] `ImportAFLSeasonPlayers` use case (AFL service) — fuzzy-match names+club; flag low-confidence; create new records for unmatched
- [ ] `just import-afl-season` CLI trigger
- [ ] AFL frontend admin page — proposed matches + new players for accept/reject

## Side quest — Pluggable FFL scoring formula *(prerequisite for Step 7a)*

- Different seasons use different scoring formulas (e.g. goals were worth 4 pts in some years, now different)
- Strategy pattern: implementations in code keyed by a string; each `ffl.season` maps to a strategy key
- Each strategy should carry a human-readable description (for frontend display)
- [ ] Design known formula variants and year ranges
- [ ] `ScoringStrategy` interface + concrete implementations
- [ ] `ffl.season.scoring_strategy` column (string key)
- [ ] Wire into score calculation use case

## Step 7 — AFL historical data import *(one-time CLI)*

- Source: afltables CSV (good coverage back to 1998)
- May be deferred to a future phase
- [ ] Evaluate afltables CSV schema; write ADR if new dependency needed
- [ ] `ImportAFLHistoricalStats` use case + CLI command
- [ ] Verify player match records, club season stats post-import

## Step 7a — FFL historical team backfill *(one-time CLI)*

- Prerequisite: side quest (pluggable scoring formula)
- Reuses `ForumPostParser` and `ImportRoundTeams` from Step 4 unchanged
- FFL scoring rules have changed over time; historical scores are imported as-recorded, not recalculated
- No `FFL.FantasyScoreCalculated` events fired on import
- May be deferred to a future phase
- [ ] Validate old forum formats work with `ForumPostParser`
- [ ] CLI command: `ParseTeamSubmission` + `ImportRoundTeams` over historical data (one round at a time)
- [ ] Verify ladder standings, scores, and player history post-import

## Close out

- [ ] Audit and remove `ffl.player.drv_name` — drop column from schema, domain, resolvers, frontend
- [ ] Retire `parse_forum.py`
- [ ] Move `afl.match.stats_import_status` + `stats_imported_at` out of core domain into `afl.dataops_match_source`
- [ ] Share `dev/postgres/test-e2e` init files with `dev/postgres/init` rather than duplicating
