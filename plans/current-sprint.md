# Current Sprint — Phase 18: Data Management — Import Infrastructure

**Sprint goal:** Build recurring data flows for team submissions, AFL stats, score reconciliation, historical backfill, and season setup. All Go; ports-and-adapters throughout; Twirp for cross-service calls.

## Cross-cutting decisions

- All Go — no Python in production; single binary deployment
- `TeamParser`, `StatsParser`, `PlayerResolver` are application-layer interfaces; adapters live in infrastructure — input source never touches use case logic
- FFL service calls AFL service via Twirp to resolve `afl_player_id` and look up players; proto definitions in `contracts/`
- `PlayerResolver` uses club code to narrow candidates before fuzzy name matching; confidence threshold gates auto-commit vs. review queue
- Frontend import feature lives in `features/data-ops/`

## Data model decisions

- `ffl.player.drv_name` is retired in principle — all Phase 18 code must not read or write it; removal (schema + domain + frontend) deferred until frontend is updated
- Player names are owned by the AFL service; the FFL service fetches them via a batch Twirp call when building a candidate pool for matching — no denormalization
- `drv_` columns elsewhere in the schema (scores, ladder) are legitimate derived/computed values and are not affected by this decision
- `ffl.player_match.afl_player_match_id` may be null at submission time (AFL stats not yet available); linked later when stats are imported
- `ffl.match` is pre-created (fixture); team submission creates/updates `ffl.club_match` and `ffl.player_match` records against it

## Tasks

- [x] ADR-018 — Twirp for cross-service communication

### Step 4 — Round team submission *(every round — implement first)*

**UI/UX decisions (agreed):**
- One team at a time
- User pre-selects FFL team + round from dropdowns before pasting
- Scores: extract player scores if present in the post; ignore position/team totals
- Low-confidence = nickname or typo; review step lets user correct before confirming

**Architecture decisions (agreed):**
- `ForumPostParser` is reused as-is for Step 7 (historical backfill — same formats, different data)
- `PlayerResolver` (fuzzy name matcher) is reused in Steps 1, 2, 4, 5; port interface takes a name, club hint, and caller-supplied candidate pool — decoupled from record type
- Step 4: candidate pool is squad-scoped (the selected FFL team's ~22 players, not the whole league)
- Step 4: candidate pool is built via a batch Twirp call to AFL service (names for squad `afl_player_id`s) — not from `ffl.player.drv_name`
- API is two-step: parse call returns parsed result + confidence scores for user review; separate confirm mutation writes `ffl.club_match` + `ffl.player_match` records to DB

- [x] Twirp: proto + buf toolchain; batch `PlayerLookup` handler on AFL service; FFL `infrastructure/rpc/` adapter (prerequisite for candidate pool)
- [x] `TeamParser` port interface (application layer)
- [x] `PlayerResolver` port interface (application layer)
- [x] `ParseTeamSubmission` use case — parse post → resolve players against squad via Twirp → return result with confidence scores (no DB writes)
- [x] `ImportRoundTeams` use case — write `ffl.club_match` + `ffl.player_match` records → fire events (confirm step)
- [x] `ForumPostParser` adapter (infrastructure) — port of `parse_forum.py`
- [x] FFL GraphQL: `parseTeamSubmission` mutation → returns parse result with confidence scores
- [x] FFL GraphQL: `confirmTeamSubmission` mutation → calls `ImportRoundTeams`
- [x] Frontend: `features/data-ops/` — club + round dropdowns, paste form, review table, confirm button
- [x] Tests: 4 parser unit tests (one per format) + 1 GraphQL integration test (Ruiboys parse+confirm) + 1 e2e golden path

### Step 5 — AFL stats import *(many times/round — automated)*

- [x] `StatsParser` port interface (application layer)
- [x] `ImportAFLStats` use case — parse stats → resolve player names via `PlayerResolver` (candidate pool = `afl.player` records for that club) → write `afl.player_match` → fire `AFL.PlayerMatchUpdated` → FFL scores recalculate
- [x] First `StatsParser` adapter for chosen data source (FootyWire scraper — `golang.org/x/net/html`, exempted from ADR by user)
- [x] `MatchSourceMapRepository` ACL table per ADR-016 (footywire match IDs cached in `afl.match_source_map`)
- [x] `MarkMatchStatsComplete` use case + `markAFLMatchStatsComplete` mutation
- [x] AFL GraphQL mutations: `importAFLMatchStats`, `markAFLMatchStatsComplete`

### Step 5a — Data ops management UI *(many times/round — manual)*

- [x] Design Data ops UI to list FFL rounds and indicate what data we have / need: FFL teams, AFL player stats
- [x] Ability to launch data imports from this page (AFL Stats Import tab with per-match scrape + mark complete)

### Step 6 — Score reconciliation *(every round)*

- `ForumPostParser` already extracts player scores (Step 4); no new parser needed
- [ ] `ReconcileScores` use case — compare imported player scores against calculated `drv_score` values; surface discrepancies
- [ ] FFL frontend — submitted scores vs calculated scores side by side; human resolves

### Step 7 — Historical backfill *(one-time per historical season — CLI)*

- Reuses `ForumPostParser` and `ImportRoundTeams` from Step 4 unchanged
- [ ] Validate old forum formats work with `ForumPostParser`
- [ ] CLI command that runs `ParseTeamSubmission` + `ImportRoundTeams` over historical data (one round at a time)

### Step 1 — AFL season player import *(once/season)*

- [ ] `ImportAFLSeasonPlayers` use case (AFL service) — fuzzy-match names+club via `PlayerResolver` (candidate pool = existing `afl.player` records); flag low-confidence; create new records for unmatched
- [ ] `just import-afl-season` CLI trigger
- [ ] AFL frontend admin page — proposed matches + new players for accept/reject

### Step 2 — FFL squad import *(once/season)*

- Twirp infrastructure (proto, AFL handler, FFL adapter) already in place from Step 4
- [ ] `ImportFLSquad` use case (FFL service) — resolve AFL player IDs via Twirp; create `ffl.player` + `ffl.player_season` records
- [ ] `just import-ffl-squad` CLI trigger
- [ ] FFL frontend admin page — proposed player mappings for accept/reject

### Step 3 — In-season player trades *(frequent)*

- [ ] FFL frontend UI for trade management
- [ ] Updates `ffl.player_season` (from/to round) via existing domain/use case layer

### Close out

- [ ] Retire `parse_forum.py`
- [ ] Remove `ffl.player.drv_name` — drop column from schema, remove from domain entity and all resolvers; update frontend to fetch player names via AFL service or search index (deferred until frontend is ready)
- [ ] Move `afl.match.stats_import_status` + `stats_imported_at` out of core domain table and into `afl.dataops_match_source` (or a dedicated `afl.dataops_match` table) — needs design: currently surfaced directly on `AFLMatch` GraphQL type; moving requires decoupling the data ops and AFL domains in the API layer
- [ ] have `dev/postgres/test-e2e` share init files, rather than having own 