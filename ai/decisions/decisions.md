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
| 008 | Accepted | — | Path-based gateway routing (/afl/query, /ffl/query); frontend uses operation-name map (not regex) |
| 009 | Accepted | ✅ | sqlc + pgx; app-layer tx via `DB.WithTx`; `MapPgError` for domain error translation |
| 010 | Accepted | — | Denormalised read models in DB; consistency maintained through domain logic on writes |
| 011 | Accepted | ✅ | Vue 3 + TypeScript (strict) + Vite + Tailwind; Apollo for server state; PrimeVue unstyled for behaviour; SPA only |
| 012 | Accepted | ✅ | Domain entities are pure and side-effect-free; repositories return fully-loaded aggregates; use cases orchestrate data loading |
| 013 | Accepted | — | No graph federation; CQRS split: GraphQL for writes/structure, search index for player stats reads |
| 014 | Accepted | ✅ | Cursor-based Connection pagination for list fields; PageInfo in common.graphqls; no offset/limit |
| 015 | Accepted | — | Typesense replaces ZincSearch; Apache 2.0, vector search path, LLM planner talks to domain not engine |