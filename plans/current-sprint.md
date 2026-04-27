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

- [ ] `TeamParser` port interface (application layer)
- [ ] `PlayerResolver` port interface (application layer)
- [ ] `ImportRoundTeams` use case — parse → resolve players → write `ffl.player_match` → fire events
- [ ] `ForumPostParser` adapter (infrastructure) — port of `parse_forum.py`
- [ ] FFL GraphQL mutation to trigger import; returns parse result with confidence scores
- [ ] Frontend: `features/data-ops/` — paste form, review table for low-confidence matches, confirm mutation
- [ ] Retire `parse_forum.py`

### Step 5 — AFL stats import *(many times/round — automated)*

- [ ] `StatsParser` port interface (application layer)
- [ ] `ImportAFLStats` use case — parse stats → write `afl.player_match` → fire `AFL.PlayerMatchUpdated` → FFL scores recalculate
- [ ] First `StatsParser` adapter for chosen data source (scrape or file)

### Step 6 — Score reconciliation *(every round)*

- [ ] `ScoreParser` port interface (application layer)
- [ ] `ForumPostParser` adapter for submitted scores
- [ ] FFL frontend — submitted scores vs calculated scores side by side; human resolves

### Step 7 — Historical backfill *(one-time per historical season — CLI)*

- [ ] Validate old forum formats work with `ForumPostParser`
- [ ] CLI command that runs `ImportRoundTeams` use case over historical data

### Step 1 — AFL season player import *(once/season)*

- [ ] `ImportAFLSeasonPlayers` use case (AFL service) — fuzzy-match names+club against `afl.player`; flag low-confidence; create new records for unmatched
- [ ] `just import-afl-season` CLI trigger
- [ ] AFL frontend admin page — proposed matches + new players for accept/reject

### Step 2 — FFL squad import *(once/season)*

- [ ] `ImportFLSquad` use case (FFL service) — resolve AFL player IDs via Twirp; create `ffl.player` + `ffl.player_season` records
- [ ] `just import-ffl-squad` CLI trigger
- [ ] FFL frontend admin page — proposed player mappings for accept/reject
- [ ] Twirp proto + `buf` toolchain (`contracts/proto/afl/v1/`, `contracts/gen/`, `just proto-gen`)
- [ ] AFL Twirp server — player lookup handler mounted at `/twirp/`
- [ ] FFL `infrastructure/rpc/` adapter implementing `PlayerLookup` port

### Step 3 — In-season player trades *(frequent)*

- [ ] FFL frontend UI for trade management
- [ ] Updates `ffl.player_season` (from/to round) via existing domain/use case layer
