# Decisions

Quick reference for agents. Read full ADRs only when you need detail on a specific decision.

| ADR | Status | Enforceable | Decision |
|-----|--------|-------------|----------|
| 001 | Accepted | ✅ | AI-optimised monorepo with `ai/` as human-agent interface |
| 002 | Accepted | ✅ | AFL and FFL expose GraphQL (gqlgen); Search exposes REST |
| 003 | Accepted | ✅ | Single PG database (`xffl`), schema isolation (`afl.*`, `ffl.*`); no cross-schema joins |
| 004 | Accepted | ✅ | PG LISTEN/NOTIFY for events; full JSON payloads; `EventDispatcher` interface in `shared/events/` |
| 005 | Accepted | ✅ | Four layers per service: Domain → Application → Infrastructure → Interface; dependencies point inward |
| 006 | Accepted | — | Zinc as search engine via dedicated Search service with event-driven indexing |
| 007 | Accepted | ✅ | `go.work` at repo root referencing all modules |
| 008 | Accepted | — | Simple reverse proxy gateway on :8090; revisit routing when second service arrives |
| 009 | Accepted | ✅ | sqlc + pgx; app-layer tx via `DB.WithTx`; `MapPgError` for domain error translation |
| 010 | Accepted | — | Denormalised read models in DB; consistency maintained through domain logic on writes |
| 011 | Accepted | — | Vue 3 + TypeScript + Vite; Apollo Client; PrimeVue |
| 012 | Accepted | ✅ | Domain entities are pure and side-effect-free; repositories return fully-loaded aggregates; use cases orchestrate data loading |