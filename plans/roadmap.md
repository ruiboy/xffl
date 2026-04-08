# Roadmap

## Context

Rebuilding from scratch using `first-cut/` as reference. Full stack (backend + frontend). Gateway introduced early so frontends always connect through it. All frontend phases require Playwright tests.

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

## Phase 12: Search Service

**Goal:** Event-driven search indexing

- [ ] Domain layer — SearchDocument, SearchQuery, SearchResult
- [ ] Application layer — Search, IndexDocument use cases; event handlers for indexing
- [ ] Infrastructure layer — Zinc REST client, event subscriber
- [ ] Interface layer — REST API (`GET /search`, `GET /health`)
- [ ] Add search passthrough to gateway
- [ ] Tests — unit (document transformation) + integration (Zinc)

## Phase 13: Search Frontend

**Goal:** Search UI (new feature, not in first-cut)

- [ ] Search view — full-text search with filters (source, type)
- [ ] Playwright tests

## Phase 14: CQRS Player Stats Read Model

**Goal:** Move player stats reads to the search index (ADR-013)

- [ ] Expand Zinc indexing to include AFL player match stats (per-round and aggregated)
- [ ] SquadView: replace AFL GraphQL stats query with search index query
- [ ] Apply pattern to other stat-heavy views as they are built

## Phase 15: Deployment

- [ ] CI-ready (GitHub Actions or similar)
- [ ] ADR — Consider deployment options (AWS, GCP, etc)

## Future Ideas

- Fully feature the UX
- Pull AFL player stats from some source
- Mobile app
- Add start timestamps to season/round/match so ordering uses real dates instead of IDs