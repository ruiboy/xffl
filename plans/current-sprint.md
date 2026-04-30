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

- `ffl.player.drv_name` is retired in principle — all Phase 20 code must not read or write it; removal (schema + domain + frontend) deferred until frontend is ready
- Player names are owned by the AFL service; the FFL service fetches them via a batch Twirp call when building a candidate pool for matching — no denormalisation
- `drv_` columns elsewhere in the schema (scores, ladder) are legitimate derived/computed values and are not affected by this decision
- `ffl.player_match.afl_player_match_id` may be null at submission time (AFL stats not yet available); linked later when stats are imported
- `ffl.match` is pre-created (fixture); team submission creates/updates `ffl.club_match` and `ffl.player_match` records against it

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

- [ ] FFL frontend UI for trade management
- [ ] Updates `ffl.player_season` (from/to round) via existing domain/use case layer

## Step 6 — Score reconciliation *(every round)*

- [ ] `ReconcileScores` use case — compare submitted player scores against calculated `drv_score`; surface discrepancies
- [ ] FFL frontend — submitted vs calculated scores side by side; human resolves

## Step 1 — AFL season player import *(once/season)*

- [ ] `ImportAFLSeasonPlayers` use case (AFL service)
- [ ] `just import-afl-season` CLI trigger
- [ ] AFL frontend admin page — proposed matches + new players for accept/reject

## Step 2 — FFL squad import *(once/season)*

- [ ] `ImportFLSquad` use case (FFL service)
- [ ] `just import-ffl-squad` CLI trigger
- [ ] FFL frontend admin page — proposed player mappings for accept/reject

## Step 7 — Historical backfill *(one-time CLI)*

- [ ] Validate old forum formats work with `ForumPostParser`
- [ ] CLI command: `ParseTeamSubmission` + `ImportRoundTeams` over historical data
- [ ] Verify ladder standings, scores, and player history post-import
- [ ] Identify AFL historical data sources (back to 1998); assess data quality
- [ ] Evaluate whether new ADRs are needed for new sources/protocols

## Close out

- [ ] Retire `parse_forum.py`
- [ ] Remove `ffl.player.drv_name` — drop column from schema, domain, resolvers, frontend
- [ ] Move `afl.match.stats_import_status` + `stats_imported_at` out of core domain into `afl.dataops_match_source`
- [ ] Share `dev/postgres/test-e2e` init files with `dev/postgres/init` rather than duplicating
