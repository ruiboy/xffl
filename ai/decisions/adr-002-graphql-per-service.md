---
status: accepted
date: 2026-04-16
scope: interface
enforceable: true
rules:
  - "each service owns its own GraphQL schema"
  - "no shared schema files across services"
---

# ADR-002: GraphQL APIs Per Service

## Context

Each service needs an API. The frontend uses Apollo Client exclusively — all calls route through the gateway as GraphQL queries.

## Decision

All services (AFL, FFL, Search) expose GraphQL endpoints using gqlgen.

## Rationale

- **Protocol consistency:** Frontend is 100% Apollo/GraphQL. A REST endpoint for Search would introduce a second API pattern (raw `fetch`) alongside the existing Apollo client.
- **Gateway uniformity:** All service routes through the gateway follow the same proxy pattern (`/{service}/query`), same CORS config.
- **Why for Hobby:** Type-safe APIs, excellent developer experience with playground/introspection, services testable independently
- **Scale Path:** Add GraphQL Federation for unified schema, field-level routing, schema stitching
