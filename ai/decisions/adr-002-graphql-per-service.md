---
status: accepted
date: 2026-03-14
scope: interface
enforceable: true
rules:
  - "each service owns its own GraphQL schema"
  - "no shared schema files across services"
---

# ADR-002: GraphQL APIs Per Service

## Context

Each service needs an API. Search uses REST (simple query/response). AFL and FFL need richer query capabilities.

## Decision

AFL and FFL expose GraphQL endpoints using gqlgen. Search exposes REST.

## Rationale

- **Why for Hobby:** Type-safe APIs, excellent developer experience with playground/introspection, services testable independently
- **Scale Path:** Add GraphQL Federation for unified schema, field-level routing, schema stitching
