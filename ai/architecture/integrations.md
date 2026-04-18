# Integration Architecture

Integrations ingest data from external sources into the domain. This document covers the architectural pattern, repo layout, and conventions for all integrations.

## Pattern: Anti-Corruption Layer (ACL)

External systems use their own identifiers, terminology, and data shapes. The ACL translates between the external model and the internal domain model. The domain never knows about external sources.

Three rules:
1. **External IDs never appear on domain entities.** Identity mapping lives in adapter-owned tables (see below).
2. **Outbound ports are defined in the Application layer.** The domain and application layers depend only on the interface — never on the adapter.
3. **Adapters live in Infrastructure.** They implement the outbound port and own all knowledge of the external source's format, protocol, and quirks.

## Layers

```
Application layer
  StatsProvider interface          ← outbound port; the contract the adapter must satisfy

Infrastructure layer
  afltables/adapter.go             ← secondary adapter; implements StatsProvider
  afltables/parser.go              ← parses the external format
  afltables/cache.go               ← fetch/cache policy

cmd/ingest/main.go                 ← entry point; wires adapter → use case → DB
```

Dependencies point inward as always: the adapter imports the port interface, not the other way around.

## Repo layout

```
services/afl/
  internal/
    application/
      ports.go                     ← outbound port interfaces (StatsProvider, etc.)
    infrastructure/
      afltables/                   ← one package per external source
        adapter.go
        parser.go
        cache.go
  cmd/
    ingest/
      main.go                      ← CLI entry point for ingestion

services/ffl/
  internal/
    application/
      ports.go
    infrastructure/
      <source>/
  cmd/
    ingest/
      main.go
```

One package per external source under `internal/infrastructure/`. If a source is replaced, only that package changes — the port interface and use cases are unaffected.

## Identity mapping

External sources have their own IDs for entities (players, clubs, rounds). These are mapped to internal domain IDs in adapter-owned tables within the service's own schema.

```sql
-- AFL service: afl schema, owned by the afltables adapter
CREATE TABLE afl.player_source_map (
    source      TEXT NOT NULL,  -- e.g. 'afltables'
    external_id TEXT NOT NULL,  -- source's player ID
    player_id   INT  NOT NULL,  -- afl.players.id
    PRIMARY KEY (source, external_id)
);
```

Rules:
- One `*_source_map` table per entity type that needs mapping, per service.
- Domain repositories never query these tables.
- No foreign keys across schemas (ADR-003). Referential integrity is enforced in application code.
- If a player is not yet in the domain, the adapter creates them before inserting the mapping.

See ADR-016 for the decision rationale.

## Fetch and cache policy

Integrations must be good citizens of the external sources they depend on.

- **Fetch minimally** — cache aggressively; only re-fetch when data is likely to have changed.
- **Document the policy** in the adapter package — e.g. "fetched at most once per week; cache cleared on Monday".
- **No hammering on startup** — if a cache is warm, use it regardless of how the process started.
- Cache state can be stored as a timestamp in the DB or as a simple file — keep it cheap.

## Event flow

After writing to the DB, adapters trigger domain events through the normal service event path:

```
AFLTablesAdapter → afl DB writes → AFL.PlayerMatchUpdated event
                                 → FFL consumes → fantasy score calculated
                                 → Search consumes → index updated
```

The adapter does not publish events directly. It calls the application use case, which writes to the DB and dispatches events as it would for any other command.

## One-time imports (historical data)

Historical data imports follow the same adapter + use case pattern, but the entry point lives under `dev/` rather than `cmd/ingest/` — it is a migration tool, not a production binary.

```
dev/import/
  ffl_historical/
    main.go      ← one-time import script
```

Once run and verified, these can be archived or deleted.
