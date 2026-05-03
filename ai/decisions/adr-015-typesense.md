---
status: accepted
date: 2026-04-16
scope: infra
enforceable: false
supersedes: ADR-006
---

# ADR-015: Replace ZincSearch with Typesense

## Context

ADR-006 selected ZincSearch on the basis of its small footprint and claimed Elasticsearch-compatible API. During Phase 13 implementation, ZincSearch v0.4.10 was found to be fundamentally broken: `term`, `match`, `bool filter`, and `query_string` with field syntax all behave as match_all. ZincSearch is also effectively abandoned — the project rebranded as OpenObserve in 2023 and v0.4.10 (January 2024) is the final release.

This project is a POC intended to inform the architecture of other projects. Future requirements include multiple document types, vector/semantic search, and an LLM as query planner.

### Engines evaluated

| Engine | Verdict |
|---|---|
| **ZincSearch** | Rejected — abandoned, broken field-specific queries |
| **Elasticsearch** | Rejected — SSPL license (non-OSI), heavy JVM footprint |
| **OpenSearch** | Considered — Apache 2.0, ES-compatible, good scale story; rejected on operational complexity and no clear managed hosting advantage for this use case |
| **Manticore Search** | Rejected — small community, insufficient basis for informing production decisions |
| **Meilisearch** | Not evaluated in depth — proprietary API, RAM-bound at scale |
| **Typesense** | Selected — see below |

### On the LLM query planner

An initial concern was that ES DSL is better represented in LLM training data than Typesense's API. On reflection this does not hold: the LLM query planner talks to the application's `SearchQuery` domain struct, not directly to the search engine API. The engine is fully behind the `DocumentRepository` interface — the LLM never writes Typesense or ES queries, it populates a well-defined struct. Search engine choice is therefore irrelevant to the LLM integration.

## Decision

Replace ZincSearch with **Typesense** as the search engine for the Search service.

The `domain.DocumentRepository` interface is unchanged. Only the infrastructure layer (`internal/infrastructure/`) is replaced. All layers above — domain, application, REST handler, gateway, frontend — are unaffected.

## Rationale

- **Managed hosting:** Typesense Cloud is a first-class managed offering — cheap at entry scale, zero ops. OpenSearch/Elasticsearch require self-hosting or a vendor relationship to achieve the same.
- **Operational simplicity:** Single binary, ~50MB RAM in dev, trivial docker-compose setup. Integration tests start in ~5 seconds vs ~30 seconds for OpenSearch. No JVM, no heap tuning.
- **Vector search:** Native vector/hybrid search (BM25 + vector in a single query). The path to AI-native search is available without changing engines.
- **License:** Apache 2.0.

## Consequences

- **Scale ceiling:** Typesense keeps its index in RAM. Practical ceiling is tens of millions of documents per collection — sufficient for projected use. If a project exceeds this, the `DocumentRepository` interface makes a migration to OpenSearch or Elasticsearch a contained infrastructure change (one package rewrite, data migration, mapping re-evaluation — not an architectural change).
- **Query API:** Typesense has its own JSON API (not ES-compatible). This is fully abstracted behind `DocumentRepository`. No application code is affected.
- **Zinc removed:** `internal/infrastructure/zinc/` is replaced with `internal/infrastructure/typesense/`. The Zinc docker-compose service is removed.

## Search Service Boundaries

The search service is solely responsible for building and maintaining its own index. Source services (AFL, FFL) must be completely unaware that a search index exists.

- **Incremental sync** flows via domain events. Source services fire events when domain data changes; the search service subscribes and reacts. The source service has no knowledge of this.
- **Full reindex** is the search service's concern. It fetches data from source services via their existing APIs and indexes directly. Source services must not expose mutations or commands whose sole purpose is to serve the search service.
- Data import use cases in source services fire domain events as a side effect of domain changes. They must not contain search-specific logic, calls, or event names designed for a specific subscriber.

## Future path

The `SearchQuery` struct evolves in the domain layer as requirements grow:

```
SearchQuery{Q, Source, Type}                          ← today
SearchQuery{Q, Source, Type, Vector []float32}        ← vector search
SearchQuery{..., Filters map[string]any, TimeRange}   ← richer filtering
```

The LLM query planner constructs a `SearchQuery` — it never knows what executes it. Swapping Typesense for a different engine if scale demands it is a contained infrastructure change, not an architectural one.
