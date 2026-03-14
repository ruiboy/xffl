# Roadmap

## Context

Rebuilding from scratch using `first-cut/` as reference. Full stack (backend + frontend). Frontend interleaved after each corresponding service. Tech choices for ORM and event system deferred to Phase 1.

## Phase 1: Foundation

**Goal:** Dev environment + shared packages + contracts

- [ ] `dev/docker-compose.yml` — PostgreSQL, Zinc (+ NATS if chosen)
- [ ] `Makefile` — targets: `up`, `down`, `migrate`, `generate`, `test`, `lint`
- [ ] Migration tooling (golang-migrate)
- [ ] `shared/database/` — DB connection helper (ORM vs sqlc — decide here)
- [ ] `shared/events/` — event dispatcher interface + implementation (PG LISTEN/NOTIFY vs NATS — decide here)
- [ ] `shared/events/memory/` — in-memory dispatcher for testing
- [ ] `contracts/events/` — shared event type definitions (`AFL.PlayerMatchUpdated`, `FFL.FantasyScoreCalculated`)

## Phase 2: AFL Service

**Goal:** First complete service with TDD

- [ ] Domain layer — Club, Season, Round, Match, PlayerMatch entities; PlayerStats value object; PlayerMatchUpdated event; repository interfaces
- [ ] Application layer — QueryClubs, UpdatePlayerMatch use cases
- [ ] Infrastructure layer — DB repositories, event publisher
- [ ] Interface layer — GraphQL schema + gqlgen resolvers, HTTP server
- [ ] Migrations — AFL schema
- [ ] Tests — unit (domain, use cases) + integration (repository, PostgreSQL)

## Phase 3: AFL Frontend

**Goal:** Vue 3 app scaffold + AFL views

- [ ] Project setup — Vue 3 + TypeScript + Vite, Apollo Client, PrimeVue, router
- [ ] AFL Clubs view — list + detail
- [ ] Component tests (Vitest)

## Phase 4: FFL Service

**Goal:** Fantasy league with cross-service event consumption

- [ ] Domain layer — Club, Player, ClubSeason, PlayerMatch entities; FantasyScore value object; FantasyScoreCalculated event; repository interfaces
- [ ] Application layer — ManagePlayers (CRUD), QueryLadder, CalculateFantasyScore use cases; AFL event subscriber
- [ ] Infrastructure layer — DB repositories, event subscriber + publisher
- [ ] Interface layer — GraphQL schema + resolvers
- [ ] Migrations — FFL schema
- [ ] Tests — unit (scoring formula, ladder) + integration (event flow)

## Phase 5: FFL Frontend

**Goal:** FFL views added to existing frontend

- [ ] FFL Players view — full CRUD
- [ ] FFL Ladder view — standings table
- [ ] Component tests

## Phase 6: Search Service

**Goal:** Event-driven search indexing

- [ ] Domain layer — SearchDocument, SearchQuery, SearchResult
- [ ] Application layer — Search, IndexDocument use cases; event handlers for indexing
- [ ] Infrastructure layer — Zinc REST client, event subscriber
- [ ] Interface layer — REST API (`GET /search`, `GET /health`)
- [ ] Tests — unit (document transformation) + integration (Zinc)

## Phase 7: Search Frontend

**Goal:** Search UI (new feature, not in first-cut)

- [ ] Search view — full-text search with filters (source, type)
- [ ] Component tests

## Phase 8: Gateway

**Goal:** Unified API entry point

- [ ] GraphQL proxy routing to AFL/FFL services
- [ ] Search passthrough to Search service
- [ ] CORS configuration + health checks
- [ ] Refactor frontend Apollo Client to point at gateway (:8090)

## Phase 9: Integration & Polish

- [ ] End-to-end tests (`tests/`)
- [ ] `make test-e2e`
- [ ] README
- [ ] CI-ready (GitHub Actions or similar)
