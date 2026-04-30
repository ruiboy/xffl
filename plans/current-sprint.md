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
