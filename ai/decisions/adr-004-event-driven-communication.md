# ADR-004: Event-Driven Cross-Service Communication

**Status:** Accepted
**Date:** 2026-03-15

## Context

AFL publishes player match stats, FFL needs them for fantasy scores, Search needs them for indexing. Services must communicate without direct coupling.

## Decision

Domain events for all cross-service communication. An `EventDispatcher` interface abstracts the transport.

**Transport: PG LISTEN/NOTIFY** for the concrete implementation.

### EventDispatcher Interface

The `EventDispatcher` interface lives in `shared/events/`. Concrete implementations:

| Implementation | Location | Purpose |
|---|---|---|
| PG LISTEN/NOTIFY | `shared/events/pg/` | Production transport |
| In-memory | `shared/events/memory/` | Testing |

### Event Payload Convention

Events carry **metadata only** (event type + aggregate ID), not full payloads. Subscribers fetch what they need. This keeps payloads well within PG's 8KB NOTIFY limit and avoids coupling subscribers to producer schemas.

### Event Flow

```
AFL.PlayerMatchUpdated → FFL (scoring) + Search (indexing)
FFL.FantasyScoreCalculated → Search (indexing)
```

## Rationale

- **No new infrastructure** — already running Postgres, so LISTEN/NOTIFY is free to operate.
- **Low event volume** — two event types, three subscribers. Purpose-built messaging (NATS, Kafka) is overkill.
- **Lost messages are tolerable** — fantasy scores can be recalculated, search can be re-indexed. Nothing requires exactly-once delivery.
- **Loose coupling** — services depend on event contracts, not on each other.
- **Testable** — in-memory dispatcher for unit/integration tests.

## Scale Path

The `EventDispatcher` interface makes the transport swappable. If LISTEN/NOTIFY becomes insufficient:

1. **Outbox table** — write events to a PG table in the same transaction as the domain change, poll and dispatch via LISTEN/NOTIFY. Adds durability without new infrastructure.
2. **NATS/JetStream** — swap the implementation for persistence, replay, consumer groups, and backpressure.

Application code does not change in either case.