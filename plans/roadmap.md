# Roadmap

## Context

Full stack rebuild (backend + frontend). Gateway introduced early so frontends always connect through it. All frontend phases require Playwright tests.

## Phase 1: Foundation ✅

**Goal:** Dev environment + shared packages + contracts

- [x] `dev/docker-compose.yml` — PostgreSQL, Zinc
- [x] `justfile` — recipes: `dev-up`, `dev-down`, `dev-reset`, `dev-seed`
- [~] Migration tooling — keeping raw SQL init scripts for now
- [x] `shared/database/` — DB connection helper
- [x] `shared/events/` — event dispatcher interface + PG LISTEN/NOTIFY implementation
- [x] `shared/events/memory/` — in-memory dispatcher for testing
- [x] `contracts/events/` — shared event type definitions

## Phase 2: AFL Service

**Goal:** Complete AFL service with sqlc

- [x] Domain layer — entities + repository interfaces
- [x] Migrate infrastructure to sqlc (ADR-009 compliance)
- [x] Application layer — mutations with `DB.WithTx` transaction support
- [x] Interface layer — GraphQL schema + resolvers + HTTP server
- [x] Tests — unit (domain) + integration (GraphQL with real DB)

## Phase 3: UX Scaffold ✅

**Goal:** Gateway + Vue 3 app scaffold + first AFL view with edit capability

- [x] ADR-008 — gateway as simple reverse proxy
- [x] ADR-011 — frontend stack (Vue 3, Apollo, Tailwind, PrimeVue unstyled)
- [x] Gateway — GraphQL proxy routing to AFL, CORS, health checks
- [x] Vue 3 project setup — TypeScript, Vite, Apollo Client (pointing at gateway :8090), router, Tailwind
- [x] AFL Match view — match result with player stats, inline editing, mutations via Apollo
- [x] Playwright e2e tests for match view (read + edit, 6 tests)

## Phase 4: AFL Frontend ✅

**Goal:** Remaining AFL views + UX polish

- [x] AI plans & prompt improvements
- [x] AFL frontend page discovery — interview, document page inventory, confirm scope
- [x] Add `aflLatestRound` backend query + `season` field on `AFLRound`
- [x] Build pages — Home (ladder + matches + round nav), Round (matches + top players + round nav), Match (read-only with club logos), Admin Match (editable player stats)
- [x] Playwright e2e tests (16 tests across home, round, match, admin-match)
- [x] UX polish — light theme, AFL club logos (18 teams), semantic Tailwind v4 `@theme` tokens, light/dark theme switcher
- [~] PrimeVue unstyled — deferred until complex interactive components are needed (Phase 5/6)

## Phase 5: FFL Service

**Goal:** FFL service as standalone CRUD with position-based fantasy scoring

- [x] Domain layer — League, Season, Round, Match, ClubSeason, ClubMatch, Player, PlayerSeason, PlayerMatch entities; position-based scoring (goals/kicks/handballs/marks/tackles/hitouts/star); bench + interchange substitution logic; repository interfaces
- [x] Application layer — ManagePlayers (CRUD), QueryLadder, CalculateFantasyScore use cases
- [x] Infrastructure layer — sqlc queries, DB repositories, transaction manager
- [x] Interface layer — GraphQL schema + resolvers
- [x] Add FFL routing to gateway
- [x] Tests — unit (scoring by position, percentage, substitution) + integration (GraphQL with real DB)

## Phase 6: FFL Frontend

**Goal:** FFL views added to existing frontend

- [x] FFL Players view
- [x] Playwright tests

## Phase 7: Data Model Refinements ✅

**Goal:** Clean up AFL/FFL data models and propagate changes through the stack

- [x] Remove `afl.player.club_id` and `ffl.player.club_id` — players exist independently; they attach to clubs for a season via `player_season → club_season`
- [x] Add `from_round_id`/`to_round_id` to `afl.player_season` for mid-season transfers
- [x] Add `status` (named/played/dnp) to `afl.player_match`
- [x] Rename `ffl.player.name → drv_name`, `ffl.player_match.score → drv_score`; domain entities keep clean names
- [x] Add unique constraint on `ffl.club_match(club_season_id, match_id)`
- [x] Drop `AFLClub.players` GraphQL field
- [x] Update domain entities, GraphQL schemas/resolvers, frontend, seed data, and tests
- [x] All 43 Playwright e2e tests green; full Go test suite green

## Phase 8: FFL UX Refinements

**Goal:** Iterative FFL frontend improvements, driven by user requests each session

- [x] Rename Roster → Squad throughout
- [x] FFL pages routed under `/ffl`; nav + home link updated; `/` redirects to `/ffl`
- [x] Architecture: no graph federation; CQRS read/write split decided (ADR-013); gateway routing clarified (ADR-008)
- [x] Fix wasteful squad query — `fflClubSeason(seasonId, clubId)` resolver; connection pagination
- [x] Fix fragile Apollo routing link — explicit operation-name map replaces regex
- [x] Global FFL club state (`useFflState`) + unified nav with club selector
- [x] Home/round page layout: circle round selector (filled/ring/ladder icon), inline headings, no matches on home
- [x] FFL eagle logo in nav (hover scales 3×)
- [x] Settings cog dropdown with dark mode toggle (cookie-persisted)
- [x] Squad page: club name heading, search panel alongside player list, Manage/Done pattern
- [x] Team Builder: club name heading, Manage/Done pattern (Done saves team)
- [x] `FFLClubSeason.season` field added to GraphQL schema and resolver

## Phase 9: FFL Team Composition Rules ✅

**Goal:** Define and enforce rules for how an FFL team is structured each round

- [x] Clarify team composition rules (positions, required structure, bench/interchange constraints)
- [x] Domain logic + validation (`ValidateTeam`, multi-starter `Score()` fix)
- [x] Enforce validation in `SetTeam` command + GraphQL mutation
- [x] Team Builder UI rebuilt with structured position layout + bench/interchange management
- [x] Domain unit tests + GraphQL integration tests

## Phase 10: Test Stabilisation ✅

**Goal:** Refactor tests to be accurate, extensible, and grounded in minimal seed data

- [x] Migrate AFL + FFL integration tests from dev Postgres to testcontainers (hermetic, no shared state)
- [x] `TestMain` + shared container pool pattern; per-test `t.Cleanup` truncates
- [x] Refactor all tests to testify (`require`/`assert`) with sentence-style `t.Run` names
- [x] Add `ai/architecture/testing.md` conventions doc + `/write-tests` skill
- [x] Delete mock-based `commands_test.go` — coverage consolidated into integration tests

## Phase 11: FFL Event Integration ✅

**Goal:** Wire up cross-service event flow between AFL and FFL

- [x] Contract extended: `RoundID` added to `PlayerMatchUpdatedPayload`
- [x] AFL publishes `AFL.PlayerMatchUpdated` after stat updates (PG LISTEN/NOTIFY)
- [x] FFL round correlation: `afl_round_id` column + join query for player_match lookup
- [x] FFL subscribes to `AFL.PlayerMatchUpdated` → auto-calculates fantasy scores
- [x] FFL publishes `FFL.FantasyScoreCalculated`
- [x] Tests — integration (event flow end-to-end, unknown player, multiple clubs)

## Phase 12: Live Round ✅

**Goal:** Compute and expose a contextually relevant "live round" across AFL and FFL, drive round nav defaults and indicators from it, and make the whole thing testable without real-time dependency.

- [x] Injectable `Clock` interface (shared); `CLOCK_OVERRIDE` env var for e2e
- [x] AFL `LiveRound` use case — `FindNeighbours` DB query + Adelaide midnight boundary; nullable (nil before season starts); no Open/Closed status — single ring style
- [x] AFL `aflLiveRound` GraphQL query (nullable)
- [x] FFL maps AFL→FFL round client-side via `afl_round_id`; no FFL service-side live round query needed
- [x] E2e seed data with fixed `start_dt` values; `CLOCK_OVERRIDE` wired into Playwright config
- [x] Frontend: `useAflState` + refactored `useFflState` — JSON cookies `xffl_afl` / `xffl_ffl` with `{ seasonId, roundId, startDate }`
- [x] RoundNav: `liveRoundId` from cookie; single `ring-active` indicator (no closed/open distinction)

## Phase 13: Search Service ✅

**Goal:** Event-driven search indexing

- [x] Domain layer — SearchDocument, SearchQuery, SearchResult
- [x] Application layer — Search, IndexDocument use cases; event handlers for indexing
- [x] Infrastructure layer — Typesense client + repository (testcontainers integration tests)
- [x] Interface layer — GraphQL (`search` query, playground, health)
- [x] Add search passthrough to gateway (`/search/query`)
- [x] Tests — unit (payload→document transformation) + integration (Typesense round-trip, filtering, upsert)

## Phase 14: Historical AFL Data — afltables (2024–present) ✅

**Goal:** Load real AFL player match stats into the domain for 2024 onward using afltables.com. Establishes the canonical player roster.

- [x] Generate `dev/postgres/seed/03_afl_historical.sql` directly from afltables CSV files (810 players, 21942 player_match rows)
- [x] Seed FFL Ruiboys squad (30 players) in `04_ffl_players.sql`
- [x] Load into dev DB; verify player/match counts and spot-check stats

## Phase 15: Database Backup ✅

**Goal:** Persist DB state durably outside the dev lifecycle.

- [x] `just backup-db` — `pg_dump --data-only | gzip` → timestamped local file; uploads via rclone if `BACKUP_REMOTE` set
- [x] `just restore-db` — reset dev DB, restore from backup, verify row counts
- [x] Restore verified end-to-end (810 players, 21942 player_match, 30 ffl players)

## Phase 16: 2026 FFL Data Import ✅

**Goal:** Seed real 2026 FFL data for rounds 1–5 (R6 squads only, no scores yet).

- [x] Identify data source — Tapatalk forum posts (manual copy-paste)
- [x] Build `dev/import/ffl/parse_forum.py` — parses all 4 team formats → `*_teams.csv` + `*_scores.csv`
- [x] Parse and validate R1–R6 (88 player rows/round); R6 squads parsed (no scores)
- [x] `resolve_squads.py` + `import_round_teams.py` — stopgap Python importers; seed SQL 04–06 generated with name-based subqueries (no hardcoded IDs)
- [x] 120 FFL players + 4 mid-season trades seeded; R1–R5 player_match + scores inserted
- [x] `dev-seed` runs all 6 seed files end-to-end (idempotent)
- [x] Verify ladder standings and scores post-import

## Phase 17: UX Improvements ✅

**Goal:** Richer, faster frontend — more detailed stats, new player and team pages, and meaningful performance improvements as data volume grows.

- [x] Performance: break up monolithic GraphQL queries — FFL (done); AFL RoundView + MatchView (done); N+1 batch fix for AFL/FFL PlayerMatches + Ladder resolvers (done); query count logging via pgx tracer (done)
- [x] DataLoader pattern (ADR-017) — per-request `Loaders` struct injected via context; `vikstrous/dataloadgen`
- [x] FFL Team Builder UX — player scores alongside names, position group totals, grand total in team summary bar, status badges, scoring formula for multiplier positions (`utils/scoring.ts`)
- [x] AFL MatchView — Manage button icon
- [x] E2e test isolation — per-test DB reset via auto worker fixture (`fixtures.ts`); `TRUNCATE … RESTART IDENTITY CASCADE` for stable IDs; `helpers/reset-db.ts` via `docker exec`; `workers: 1`; restored `ffl-team-builder.spec.ts`
- [x] E2e docs — `ai/architecture/testing.md` Playwright section; cookbook recipe for adding new specs

## Phase 18: Data Management — Import Infrastructure, Part I

**Goal:** Build recurring data flows for team submissions, AFL stats, score reconciliation, historical backfill, and season setup. All Go; ports-and-adapters throughout; Twirp for cross-service calls.

- [x] ADR — Twirp for cross-service communication
- [x] FFL Round team submission
- [x] AFL stats import

## Phase 19: Graph Federation

**Goal:** Adopt Apollo Federation so the frontend can traverse cross-service entity relationships in a single query. Replace the path-based gateway with Apollo Router. Establish `AFLPlayerSeason` as a first-class graph type.

- [x] Apollo Router — replace `services/gateway` with Apollo Router in Docker Compose; configure supergraph from both subgraphs
- [x] AFL subgraph — add `AFLPlayerSeason` type + entity resolver; add `@key` to `AFLPlayerMatch`; mount federation-compatible handler
- [x] FFL subgraph — add `aflPlayerSeason` field on `FFLPlayerSeason`; add `aflPlayerMatch` field on `FFLPlayerMatch`; reference resolvers; mount federation-compatible handler
- [x] Frontend — single Apollo client endpoint; remove operation-name routing link
- [x] Tests + e2e verification

## Phase 20: Data Management — Import Infrastructure, Part II

- [ ] FFL In-season player trades
- [ ] FFL Score reconciliation
- [ ] AFL season player import
- [ ] FFL squad import
- [ ] FFL Historical backfill
- [ ] Close out

See detailed breakdown (carried over from Phase 18) in [phase-20-sprint.md](phase-20-sprint.md)

## Phase 21: Search Frontend + Index Enrichment

**Goal:** Search UI backed by an enriched Typesense index with whatever data the UX needs.

- [ ] Search view — full-text search with filters (source, type)
- [ ] Expand search index as needed to support UX data requirements (player stats, aggregates, etc.)
- [ ] Playwright tests

## Phase 22: UX — Player & Team Pages

**Goal:** Player and team detail pages with career stats, season history, and richer stat data across existing views.

- [ ] Player pages — career stats, season history, club timeline
- [ ] Team pages — squad, round-by-round scores, season summary
- [ ] Other season pages (TBD based on usage)
- [ ] Richer stat data surfaced in existing views

## Phase 23: CQRS Player Stats Read Model

**Goal:** Move player stats reads to the search index (ADR-013)

- [ ] Expand Typesense indexing to include AFL player match stats (per-round and aggregated)
- [ ] SquadView: replace AFL GraphQL stats query with search index query
- [ ] Apply pattern to other stat-heavy views as they are built

## Phase 24: Deployment

- [ ] CI-ready (GitHub Actions or similar)
- [ ] ADR — Consider deployment options (AWS, GCP, etc)

## Future Ideas

- Fully feature the UX
- Live AFL data source for ongoing weekly stats (TBD — afltables may serve for weekly reconciliation once historical load is complete)
- Mobile app
- Backup remote destination — choose cloud storage (recommend rclone: supports S3, GCS, B2 with a single consistent interface). Decide when deployment target is clearer.