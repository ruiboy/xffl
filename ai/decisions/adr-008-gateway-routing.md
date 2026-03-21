---
status: deferred
date: 2026-03-14
scope: interface
enforceable: false
---

# ADR-008: Gateway Routing Strategy

## Context

Gateway proxies frontend requests to the correct backend service. First-cut used simple string-based routing (presence of "afl"/"ffl" in query text). Works but fragile.

## Options

1. **String-based routing** — zero config, predictable, breaks if query names don't contain service identifiers
2. **GraphQL query parsing** — route by field/type, more robust
3. **GraphQL Federation** — most capable, most complex

## Decision

Deferred to Phase 8. Frontend connects directly to services during earlier phases. Decision informed by real query patterns.
