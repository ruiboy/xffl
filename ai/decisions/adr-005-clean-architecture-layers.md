# ADR-005: Clean Architecture with Go Idioms

**Status:** Accepted
**Date:** 2026-03-14

## Context

Services need a consistent internal structure that enforces separation of concerns and testability.

## Decision

Four layers per service: Domain → Application → Infrastructure → Interface. Dependencies point inward. Domain entities are pure structs. Interfaces defined where consumed.

## Rationale

- **Why for Hobby:** Enforces good separation, domain logic testable with plain Go, educational value, makes codebase navigable
- **Scale Path:** Patterns scale well — can add CQRS, event sourcing, or simplify to layered architecture as needed
