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

Option 1: simple reverse proxy. Only AFL exists today — no routing decision required. When FFL arrives (Phase 5), upgrade to option 2 or 3 based on real query patterns.

Gateway runs on `:8090`, proxies `/query` to AFL (`:8080`), handles CORS, and exposes `/health`.
