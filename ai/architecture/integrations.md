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
  <Port> interface                 ← outbound port; the contract the adapter must satisfy

Infrastructure layer
  <source>/adapter.go              ← secondary adapter; implements the port
  <source>/...                     ← parser, cache, etc.

cmd/ingest/main.go                 ← entry point; wires adapter → use case → DB
```

Dependencies point inward as always: the adapter imports the port interface, not the other way around.

## Repo layout

```
services/{afl,ffl}/
  internal/
    application/
      ports.go                     ← outbound port interfaces
    infrastructure/
      <source>/                    ← one package per external source
        doc.go                     ← system description, role, cache policy
        adapter.go
        ...
  cmd/
    ingest/
      main.go                      ← CLI entry point

```

One package per external source under `internal/infrastructure/`. If a source is replaced, only that package changes — the port interface and use cases are unaffected.

## Adapter directory convention

Every adapter package must contain a `doc.go` file with a package-level comment covering:
- What external system this adapter talks to
- Its role in the data pipeline (what it ingests and where it writes)
- The fetch/cache policy in plain language

Using `doc.go` (rather than a README) means the description is surfaced by `go doc`, shown in IDE hover tooltips, and lives where the code is. It is the first thing a developer reads when landing in an unfamiliar adapter package.

```go
// Package <source> fetches <description> and writes into the domain via the <Port> port.
//
// Cache policy: <fetch frequency and invalidation rule>.
package <source>
```

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

