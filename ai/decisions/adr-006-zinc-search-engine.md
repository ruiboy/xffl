---
status: superseded
date: 2026-03-14
superseded-by: ADR-015
scope: infra
enforceable: false
---

# ADR-006: Zinc Search Engine

## Context

Need full-text search across players, clubs, and matches from both AFL and FFL. Options: PostgreSQL FTS, Elasticsearch, Meilisearch, Typesense, Zinc.

## Decision

Zinc as a dedicated search engine, accessed via a separate Search service with event-driven indexing.

## Rationale

- **Why for Hobby:** Single Go binary (~40MB), ~100MB RAM, Elasticsearch-compatible API, better relevance/faceting than PG FTS, native Go ecosystem fit
- **Scale Path:** Direct migration to Elasticsearch/OpenSearch — existing queries and index mappings remain compatible
