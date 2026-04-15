# Current Sprint — Phase 13: Search Service

**Sprint goal:** Build a standalone Search service that subscribes to AFL and FFL events, indexes documents into Typesense, and exposes a GraphQL API for full-text search with source/type filtering.

> **Pivot notes (2026-04-16):**
> - ADR-015 replaced ZincSearch with Typesense. Tasks 1-3 unaffected. Task 4 redone.
> - ADR-002 updated: Search now exposes GraphQL (like AFL/FFL) instead of REST, keeping the frontend on a single protocol (Apollo). Task 5 redone.

## Design decisions

- **GraphQL, not REST** (ADR-002 updated) — `search(q, source, type)` query via gqlgen, consistent with AFL/FFL. Frontend stays 100% Apollo.
- **Single Typesense collection** (`documents`) — all documents in one collection with `source` and `type` fields for filtering. Simpler than per-type collections; supports the cross-source search the frontend will need.
- **Document ID** — `"{source}_{type}_{id}"` (e.g. `"afl_player_match_42"`). Deterministic so re-indexing is idempotent (Typesense upserts by `id`).
- **Event subscriptions** — same pattern as FFL: `dispatcher.Subscribe(...)` then `go dispatcher.Listen(ctx)`. Subscribes to both `AFL.PlayerMatchUpdated` and `FFL.FantasyScoreCalculated`.
- **No DB** — Search service has no PostgreSQL schema. Typesense is its only persistence.
- **Typesense auth** — API key set via `TYPESENSE_API_KEY` env var (default `xyz` for local dev, matching docker-compose).
- **Testing** — unit tests for payload→document transformation; integration tests against a real Typesense instance via testcontainers.
- **Gateway passthrough** — `/search/query` proxied to Search service (same pattern as `/afl/query`, `/ffl/query`).

## Document shapes

### AFL player match (from `AFL.PlayerMatchUpdated`)
```
source: "afl"
type:   "player_match"
id:     player_match_id
fields: player_match_id, player_season_id, club_match_id, round_id,
        kicks, handballs, marks, hitouts, tackles, goals, behinds
```

### FFL fantasy score (from `FFL.FantasyScoreCalculated`)
```
source: "ffl"
type:   "fantasy_score"
id:     player_match_id (FFL)
fields: player_match_id, score, afl_player_match_id
```

## Tasks

### 1. Scaffold
- [x] Create `services/search/` directory structure: `cmd/`, `internal/domain/`, `internal/application/`, `internal/infrastructure/zinc/`, `internal/interface/rest/`
- [x] `services/search/go.mod` — module `xffl/services/search`; import `xffl/contracts/events`, `xffl/shared/events`
- [x] Add `./services/search` to `go.work`

### 2. Domain layer
- [x] `internal/domain/document.go` — `SearchDocument{ID, Source, Type, Data map[string]any}` struct; `Source` and `Type` string constants (`SourceAFL`, `SourceFFL`, `TypePlayerMatch`, `TypeFantasyScore`)
- [x] `internal/domain/query.go` — `SearchQuery{Q, Source, Type string}` and `SearchResult{Total int, Documents []SearchDocument}` structs
- [x] `internal/domain/repository.go` — `DocumentRepository` interface: `Index(ctx, doc) error` and `Search(ctx, query) (SearchResult, error)`

### 3. Application layer
- [x] `internal/application/index.go` — `IndexDocument` use case: takes a `SearchDocument`, delegates to `DocumentRepository.Index`
- [x] `internal/application/search.go` — `Search` use case: takes a `SearchQuery`, delegates to `DocumentRepository.Search`; returns `SearchResult`
- [x] `internal/application/handlers.go` — `HandlePlayerMatchUpdated(ctx, payload []byte) error` and `HandleFantasyScoreCalculated(ctx, payload []byte) error`; each unmarshal contract payload, build `SearchDocument`, call `IndexDocument`
- [x] Unit tests for handler payload→document transformation (table-driven: valid payload, malformed JSON, zero values)

### 4. Infrastructure: Typesense client *(redo — was Zinc, see ADR-015)*
- [x] `internal/infrastructure/typesense/client.go` — `Client{apiURL, apiKey, collection, httpClient}`; `EnsureCollection(ctx) error` creates the `documents` collection with schema (`source`/`type` as string facets, `.*` auto for data fields); `upsertDoc(ctx, doc) error`; `search(ctx, q, queryBy, filterBy) (*searchResponse, error)`
- [x] `internal/infrastructure/typesense/repository.go` — wraps `Client`, implements `domain.DocumentRepository`; maps Typesense response to `SearchResult`; uses native `filter_by` (no post-filtering needed)
- [x] Integration tests — use testcontainers (Typesense image: `typesense/typesense:27.1`) to start Typesense; test `Index` then `Search` round-trip; test source/type filtering; test idempotent re-index via upsert
- [x] Remove `internal/infrastructure/zinc/` directory

### 5. Interface: GraphQL *(redo — was REST, see ADR-002 update)*
- [x] `api/graphql/schema.graphqls` — `search(q: String, source: String, type: String): SearchResult!` query; `SearchResult{total, documents}` and `SearchDocument{id, source, type, data: JSON}` types
- [x] `gqlgen.yml` + `go generate` — generated resolver scaffold, models, executor
- [x] `internal/interface/graphql/resolver.go` — `Resolver{Repo domain.DocumentRepository}`; wire search query to repo
- [x] Unit tests for resolver (stub repo, verify query mapping and response shape)
- [x] Remove `internal/interface/rest/` (replaced by GraphQL)

### 6. Service entrypoint *(update for GraphQL)*
- [x] `cmd/main.go` — wire `TypesenseClient → TypesenseRepository → IndexDocument + Handlers + GraphQL server`; read `TYPESENSE_HOST` (default `localhost`), `TYPESENSE_PORT` (default `8108`), `TYPESENSE_API_KEY` (default `xyz`); call `EnsureCollection` on startup; subscribe to both events via pgevents; serve `/query` (GraphQL) + `/` (playground) + `/health`; port 8082

### 7. Gateway: search passthrough
- [x] `services/gateway/cmd/main.go` — add `SEARCH_SERVICE_URL` env (default `http://localhost:8082`); add `/search/query` route (same pattern as `/afl/query`, `/ffl/query`)

### 8. justfile + docker-compose + go.work
- [x] `dev/docker-compose.yml` — removed `zinc` service, added `typesense` service (image `typesense/typesense:27.1`, port 8108, API key `xyz`, data volume)
- [x] `justfile` — updated comments/echo to reference Typesense; added `run-search` to `run-all` and port 8082 to `stop-all`
- [x] `go.work` — already has `./services/search` from task 1
