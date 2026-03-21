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

## Phase 3: UX Scaffold

**Goal:** Gateway + Vue 3 app scaffold + first AFL view with edit capability

- [ ] ADR — choose design system / component library
- [ ] Gateway — GraphQL proxy routing to AFL, CORS, health checks
- [ ] Vue 3 project setup — TypeScript, Vite, Apollo Client (pointing at gateway :8090), router
- [ ] AFL Match view — match result with player stats, inline editing of player stats
- [ ] Playwright tests for match view (read + edit)

## Phase 4: AFL Frontend

**Goal:** Remaining AFL views

- [ ] AFL Clubs view — list + detail
- [ ] AFL Season view — ladder, rounds, matches
- [ ] Playwright tests

## Phase 5: FFL Service

**Goal:** Fantasy league with cross-service event consumption

- [ ] Domain layer — Club, Player, ClubSeason, PlayerMatch entities; FantasyScore value object; FantasyScoreCalculated event; repository interfaces
- [ ] Application layer — ManagePlayers (CRUD), QueryLadder, CalculateFantasyScore use cases; AFL event subscriber
- [ ] Infrastructure layer — DB repositories, event subscriber + publisher
- [ ] Interface layer — GraphQL schema + resolvers
- [ ] Add FFL routing to gateway
- [ ] Tests — unit (scoring formula, ladder) + integration (event flow)

## Phase 6: FFL Frontend

**Goal:** FFL views added to existing frontend

- [ ] FFL Players view — full CRUD
- [ ] FFL Ladder view — standings table
- [ ] Playwright tests

## Phase 7: Search Service

**Goal:** Event-driven search indexing

- [ ] Domain layer — SearchDocument, SearchQuery, SearchResult
- [ ] Application layer — Search, IndexDocument use cases; event handlers for indexing
- [ ] Infrastructure layer — Zinc REST client, event subscriber
- [ ] Interface layer — REST API (`GET /search`, `GET /health`)
- [ ] Add search passthrough to gateway
- [ ] Tests — unit (document transformation) + integration (Zinc)

## Phase 8: Search Frontend

**Goal:** Search UI (new feature, not in first-cut)

- [ ] Search view — full-text search with filters (source, type)
- [ ] Playwright tests

## Phase 9: Deployment

- [ ] CI-ready (GitHub Actions or similar)
- [ ] ADR - Consider deployment options - AWS, GCP, etc

## Future Ideas

- Fully feature the UX
- Pull AFL player stats from some source
- Mobile app