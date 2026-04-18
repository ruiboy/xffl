---
status: accepted
date: 2026-04-18
scope: infra
enforceable: true
---

# ADR-016: `integrations` Schema for External Identity Mapping (ACL)

## Context

Integrations require ingesting data from external sources (e.g. AFL stats from web sources.). Each external source uses its own identifiers for entities (players, clubs, rounds) that must be mapped to internal domain IDs.

Storing these external IDs directly on domain entities (e.g. `afl.players.source_system_id`) would leak integration concerns into the domain model. The AFL domain should have no knowledge of external sources.

In DDD terms, the correct pattern is an **Anti-Corruption Layer (ACL)**: a translation layer that maps external identities and concepts to internal domain identities. The mapping table is infrastructure owned by the ACL, not the domain.

The question is where this infrastructure lives in the database.

### Options considered

| Option | Assessment |
|--------|-----------|
| Columns on domain tables (e.g. `afl.players.source_system_id`) | Rejected — leaks integration concerns into the domain schema |
| Tables in the `afl` / `ffl` schema (e.g. `afl.player_source_map`) | Rejected — pollutes domain schemas with ACL plumbing; harder to reason about at a glance |
| Dedicated `integrations` schema | Selected — see below |

## Decision

Add an **`integrations`** PostgreSQL schema to the shared database. All external identity mapping tables live here, owned by the integration adapters (infrastructure layer), never referenced by domain queries.

Initial table:

```sql
CREATE TABLE integrations.player_source_map (
    source      TEXT NOT NULL,  -- e.g. 'web-resource-1'
    external_id TEXT NOT NULL,  -- source's identifier
    player_id   INT  NOT NULL,  -- afl.players.id
    PRIMARY KEY (source, external_id)
);
```

Additional mapping tables are added here as new sources and entity types require them (clubs, rounds, FFL teams, etc.).

## Rationale

- **ACL concerns are isolated** — domain schemas (`afl.*`, `ffl.*`) remain free of integration plumbing.
- **Scales naturally** — as more sources and entity types are integrated, the pattern is consistent and the location is obvious.
- **Consistent with ADR-003** — ADR-003 established schema isolation as the boundary mechanism. The `integrations` schema follows the same principle: integration infrastructure owns its schema, domain services own theirs. No cross-schema joins from domain queries.

## Consequences

- Integration adapters read/write `integrations.*` tables directly; domain repositories never reference them.
- The `integrations` schema is initialised in `dev/postgres/init/` alongside `01_afl_schema.sql` and `02_ffl_schema.sql`.
- No foreign keys across schemas (consistent with ADR-003). Referential integrity for `player_id` is enforced in application code.
