# Current Sprint — Phase 13: Search Service

**Sprint goal:** Build a standalone Search service that subscribes to AFL and FFL events, indexes documents into Typesense, and exposes a REST API for full-text search with source/type filtering.

> **Pivot note (2026-04-16):** ADR-015 replaced ZincSearch with Typesense. Tasks 1-3 (scaffold, domain, application) are unaffected — they program to `DocumentRepository`. Task 4 (infrastructure layer) is being redone for Typesense. See ADR-015 for rationale.

## Design decisions

- **REST, not GraphQL** (ADR-002) — `GET /search?q=...&source=...&type=...` and `GET /health`.
- **Single Typesense collection** (`documents`) — all documents in one collection with `source` and `type` fields for filtering. Simpler than per-type collections; supports the cross-source search the frontend will need.
- **Document ID** — `"{source}_{type}_{id}"` (e.g. `"afl_player_match_42"`). Deterministic so re-indexing is idempotent (Typesense upserts by `id`).
- **Event subscriptions** — same pattern as FFL: `dispatcher.Subscribe(...)` then `go dispatcher.Listen(ctx)`. Subscribes to both `AFL.PlayerMatchUpdated` and `FFL.FantasyScoreCalculated`.
- **No DB** — Search service has no PostgreSQL schema. Typesense is its only persistence.
- **Typesense auth** — API key set via `TYPESENSE_API_KEY` env var (default `xyz` for local dev, matching docker-compose).
- **Testing** — unit tests for payload→document transformation; integration tests against a real Typesense instance via testcontainers.
- **Gateway passthrough** — `/search` prefix forwarded to Search service (strips prefix, proxies to port 8082).

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
- [ ] `internal/infrastructure/typesense/client.go` — `Client{host, port, apiKey, protocol, httpClient}`; `EnsureCollection(ctx) error` creates the `documents` collection with schema (fields: `id`, `source`, `type`, `data` as string fields; `source` and `type` as facet fields); `Index(ctx, doc) error` upserts document; `Search(ctx, query) (SearchResult, error)` calls Typesense search endpoint with `q`, `query_by`, `filter_by`
- [ ] `internal/infrastructure/typesense/repository.go` — wraps `Client`, implements `domain.DocumentRepository`; maps Typesense response to `SearchResult`
- [ ] Integration tests — use testcontainers (Typesense image: `typesense/typesense:27.1`) to start Typesense; test `Index` then `Search` round-trip; test source/type filtering; test idempotent re-index via upsert
- [ ] Remove `internal/infrastructure/zinc/` directory

### 5. Interface: REST handlers
- [ ] `internal/interface/rest/handler.go` — `Handler{search *application.Search}`; `ServeSearch(w, r)` reads `q`, `source`, `type` query params, calls use case, writes JSON response `{"total": N, "documents": [...]}`
- [ ] `GET /health` in `cmd/main.go` (same pattern as AFL/FFL)

### 6. Service entrypoint
- [ ] `cmd/main.go` — wire `TypesenseClient → TypesenseRepository → IndexDocument + Search + handlers`; read `TYPESENSE_HOST` (default `localhost`), `TYPESENSE_PORT` (default `8108`), `TYPESENSE_API_KEY` (default `xyz`); call `EnsureCollection` on startup; subscribe to both events; start REST server on port 8082

### 7. Gateway: search passthrough
- [ ] `services/gateway/cmd/main.go` — add `SEARCH_SERVICE_URL` env (default `http://localhost:8082`); add `/search` route that strips the `/search` prefix and proxies to Search service; add `GET` to CORS allowed methods

### 8. justfile + docker-compose + go.work
- [ ] `dev/docker-compose.yml` — remove `zinc` service, add `typesense` service (image `typesense/typesense:27.1`, port 8108, API key `xyz`, data dir `/data`)
- [ ] `justfile` — update comments/echo to reference Typesense instead of Zinc; update port from 4080 to 8108; add `run-search` to `run-all` and `stop-all` (port 8082)
- [ ] `go.work` — add `./services/search` (if not already present from task 1)
