# ADR-004: Event-Driven Cross-Service Communication

**Status:** Accepted
**Date:** 2026-03-14

## Context

AFL publishes player match stats, FFL needs them for fantasy scores, Search needs them for indexing. Services must communicate without direct coupling.

## Decision

Domain events for all cross-service communication. An `EventDispatcher` interface abstracts the transport. Concrete implementation (PG LISTEN/NOTIFY vs NATS) deferred to Phase 1.

## Rationale

- **Why for Hobby:** Loose coupling, testable with in-memory dispatcher, eventual consistency is fine for fantasy scores
- **Scale Path:** Swap transport without changing application code — interface supports migration to NATS, Redis, AWS EventBridge, Kafka

## Event Flow

```
AFL.PlayerMatchUpdated → FFL (scoring) + Search (indexing)
FFL.FantasyScoreCalculated → Search (indexing)
```
