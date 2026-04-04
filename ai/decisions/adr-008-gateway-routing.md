---
status: accepted
date: 2026-03-22
scope: interface
enforceable: false
---

# ADR-008: Gateway Routing Strategy

## Context

Gateway proxies frontend requests to the correct backend service. First-cut used simple string-based routing (presence of "afl"/"ffl" in query text). Works but fragile.

Phase 3 brings the gateway forward to provide a single entry point for the Vue frontend.

## Options

1. **Simple reverse proxy** — stdlib `httputil.ReverseProxy`, forward `/query` to AFL service. No routing logic needed while there is only one service.
2. **String-based routing** — zero config, predictable, breaks if query names don't contain service identifiers
3. **GraphQL query parsing** — route by field/type, more robust
4. **GraphQL Federation** — most capable, most complex

## Decision

**Phase 3:** Option 1 — simple reverse proxy to AFL only.

**Phase 5 (current):** Option 2 — path-based routing. Gateway exposes `/afl/query` → AFL (`:8080`) and `/ffl/query` → FFL (`:8081`). Frontend Apollo client uses a custom link that inspects operation field names to pick the correct endpoint.

**Going forward:** Stay on path-based routing. Federation (option 4) was evaluated and rejected — see ADR-013. The frontend routing link should migrate from regex-based field name matching to an explicit operation-name map to eliminate fragility.

Gateway runs on `:8090`, handles CORS, and exposes `/health`.
