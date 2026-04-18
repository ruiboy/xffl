---
status: accepted
date: 2026-04-18
scope: infra
enforceable: true
---

# ADR-016: ACL Identity Mapping Tables Within Service Schemas

## Context

Integrations require ingesting data from external sources (e.g. AFL stats from web sources). Each external source uses its own identifiers for entities (players, clubs, rounds) that must be mapped to internal domain IDs.

Storing these external IDs directly on domain entities (e.g. `afl.players.source_system_id`) would leak integration concerns into the domain model. The AFL domain should have no knowledge of external sources.

In DDD terms, the correct pattern is an **Anti-Corruption Layer (ACL)**: a translation layer that maps external identities to internal domain identities. The mapping table is infrastructure owned by the ACL adapter, not the domain.

The question is where these mapping tables live in the database.

### Options considered

| Option | Assessment |
|--------|-----------|
| Columns on domain tables (e.g. `afl.players.source_system_id`) | Rejected — leaks integration concerns into domain entities |
| Dedicated `integrations` schema shared across services | Rejected — creates a cross-service shared schema; services would access two schemas, violating ADR-003's schema-per-service boundary |
| Adapter-owned tables within the service's own schema | Selected — see below |

## Decision

ACL identity mapping tables live in the **service's own schema**, owned by the integration adapter (infrastructure layer). They are never referenced by domain repositories or queries.

Example for the AFL service:

```sql
CREATE TABLE afl.player_source_map (
    source      TEXT NOT NULL,  -- e.g. 'afltables'
    external_id TEXT NOT NULL,  -- source's identifier for the player
    player_id   INT  NOT NULL,  -- afl.players.id
    PRIMARY KEY (source, external_id)
);
```

If FFL requires similar mapping tables, they are defined in the `ffl` schema. Each service owns and manages its own ACL tables.

## Rationale

- **Schema boundaries preserved** — each service accesses only its own schema (`afl.*` or `ffl.*`), consistent with ADR-003. A shared `integrations` schema would require services to access two schemas and become a cross-service shared dependency.
- **Domain entities stay clean** — external IDs never appear on domain entity tables.
- **Duplication is acceptable** — the pattern is duplicated across services, but this is consistent with the principle of preferring duplication over incorrect abstractions (see principles.md). The tables serve different services with different mapping concerns.

## Consequences

- Integration adapters read/write their service's `*_source_map` tables directly; domain repositories never reference them.
- No foreign keys across schemas (consistent with ADR-003). Referential integrity is enforced in application code.
- Schema init SQL for mapping tables lives alongside other schema SQL in `dev/postgres/init/`.
