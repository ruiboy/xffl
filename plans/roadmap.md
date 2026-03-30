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

- [ ] Domain layer — League, Season, Round, Match, ClubSeason, ClubMatch, Player, PlayerSeason, PlayerMatch entities; position-based scoring (goals/kicks/handballs/marks/tackles/hitouts/star); bench + interchange substitution logic; repository interfaces
- [ ] Application layer — ManagePlayers (CRUD), QueryLadder, CalculateFantasyScore use cases
- [ ] Infrastructure layer — sqlc queries, DB repositories, transaction manager
- [ ] Interface layer — GraphQL schema + resolvers
- [ ] Add FFL routing to gateway
- [ ] Tests — unit (scoring by position, percentage, substitution) + integration (GraphQL with real DB)

## Phase 6: FFL Frontend

**Goal:** FFL views added to existing frontend

- [ ] FFL Players view — full CRUD
- [ ] FFL Ladder view — standings table
- [ ] Playwright tests

## Phase 7: FFL Event Integration

**Goal:** Wire up cross-service event flow between AFL and FFL

- [ ] FFL subscribes to `AFL.PlayerMatchUpdated` → auto-calculates fantasy scores
- [ ] FFL publishes `FFL.FantasyScoreCalculated`
- [ ] Tests — integration (event flow end-to-end)

## Phase 8: Search Service

**Goal:** Event-driven search indexing

- [ ] Domain layer — SearchDocument, SearchQuery, SearchResult
- [ ] Application layer — Search, IndexDocument use cases; event handlers for indexing
- [ ] Infrastructure layer — Zinc REST client, event subscriber
- [ ] Interface layer — REST API (`GET /search`, `GET /health`)
- [ ] Add search passthrough to gateway
- [ ] Tests — unit (document transformation) + integration (Zinc)

## Phase 9: Search Frontend

**Goal:** Search UI (new feature, not in first-cut)

- [ ] Search view — full-text search with filters (source, type)
- [ ] Playwright tests

## Phase 10: Deployment

- [ ] CI-ready (GitHub Actions or similar)
- [ ] ADR - Consider deployment options - AWS, GCP, etc

## Future Ideas

- Fully feature the UX
- Pull AFL player stats from some source
- Mobile app
- Add start timestamps to season/round/match so ordering uses real dates instead of IDs