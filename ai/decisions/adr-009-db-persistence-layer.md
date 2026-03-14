# ADR-009: Database Persistence Layer

**Status:** Deferred (Phase 1)
**Date:** 2026-03-14

## Context

Services need to interact with PostgreSQL. First-cut used GORM.

## Options

1. **GORM** — full ORM, struct tags, familiar from first-cut. Magic can be surprising.
2. **sqlc + pgx** — write SQL, generate type-safe Go. No magic, compile-time safety, more Go-idiomatic.
3. **Raw pgx** — maximum control, more boilerplate.

## Decision

Deferred to Phase 1 when building `shared/database/`. Either way, domain entities remain pure — DB entities and converters live in infrastructure layer (ADR-005).
