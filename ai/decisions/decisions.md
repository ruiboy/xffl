# Decisions

Quick reference for agents. Read full ADRs only when you need detail on a specific decision.

| ADR | Status | Enforceable | Decision |
|-----|--------|-------------|----------|
| 001 | Accepted | ✅ | AI-optimised monorepo with `ai/` as human-agent interface |
| 002 | Accepted | ✅ | All services expose GraphQL (gqlgen) |
| 003 | Accepted | ✅ | Single PG database (`xffl`), schema isolation (`afl.*`, `ffl.*`); no cross-schema joins |
| 004 | Accepted | ✅ | PG LISTEN/NOTIFY for events; full JSON payloads; `EventDispatcher` interface in `shared/events/` |
| 005 | Accepted | ✅ | Four layers per service: Domain → Application → Infrastructure → Interface; dependencies point inward |
| 006 | Superseded | — | ~~Zinc as search engine~~ — see ADR-015 |
| 007 | Accepted | ✅ | `go.work` at repo root referencing all modules |
| 008 | Superseded | — | ~~Path-based gateway routing~~ — see ADR-013 (rev. 2026-04-28) |
| 009 | Accepted | ✅ | sqlc + pgx; app-layer tx via `DB.WithTx`; `MapPgError` for domain error translation |
| 010 | Accepted | — | Denormalised read models in DB; consistency maintained through domain logic on writes |
| 011 | Accepted | ✅ | Vue 3 + TypeScript (strict) + Vite + Tailwind; Apollo for server state; PrimeVue unstyled for behaviour; SPA only |
| 012 | Accepted | ✅ | Domain entities are pure and side-effect-free; repositories return fully-loaded aggregates; use cases orchestrate data loading |
| 013 | Accepted | ✅ | Apollo Federation for structural entity traversal; Typesense for aggregated reads; Apollo Router replaces gateway (rev. 2026-04-28) |
| 014 | Accepted | ✅ | Cursor-based Connection pagination for list fields; PageInfo in common.graphqls; no offset/limit |
| 015 | Accepted | — | Typesense replaces ZincSearch; Apache 2.0, vector search path, LLM planner talks to domain not engine |
| 016 | Accepted | ✅ | ACL identity mapping: xref tables per source per entity (e.g. `afl.xref_<source>_player`); owned by adapters/import tools; no FK to core schema; no shared integration schema |
| 017 | Accepted | ✅ | `vikstrous/dataloadgen` as the convention for all resolver entity lookups; resolvers never call single-item repo methods by ID; per-request `Loaders` struct in context; batch functions delegate to repository `FindByIDs` |
| 018 | Accepted | ✅ | Twirp for synchronous cross-service RPC; port interface in application layer; proto in `contracts/proto/`, generated stubs in `contracts/gen/`; `buf` for codegen |