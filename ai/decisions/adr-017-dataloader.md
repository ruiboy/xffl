---
status: accepted
date: 2026-04-26
scope: interface
enforceable: true
rules:
  - "resolvers never call single-item repository methods by ID — every such lookup uses a DataLoader"
  - "loaders are initialised once per request in gqlgen AroundOperations middleware"
  - "loaders are accessed via context — never constructed inside a resolver"
  - "batch functions delegate to repository interfaces — no raw DB calls in loaders"
---

# ADR-017: DataLoader as the Convention for All Resolver Entity Lookups

## Context

GraphQL resolvers fetch related entities by ID. gqlgen executes each resolver independently, so without an explicit batching mechanism, N parent items yield N separate round-trips for the same entity type. In a schema with several join hops this compounds; over HTTP (cross-service lookups) the latency multiplier is far worse than a DB N+1.

The N+1 problem surfaces in several forms:

- **Parent-driven**: a resolver iterates a slice and calls `GetPlayer(ctx, id)` per item
- **Sibling-driven**: two independent resolvers (e.g. `HomeClubMatch`, `AwayClubMatch`) each fetch the same parent entity with no shared scope to coordinate from
- **Cross-service**: FFL views surfacing AFL player stats — N HTTP calls to the AFL service where 1 batched call suffices

These are all the same problem at different layers. The resolver-per-entity execution model requires an explicit batching mechanism at every lookup boundary.

### Upcoming pressure

- **Cross-service data**: FFL team and squad views will surface AFL player stats. Without batching, this is an N+1 over HTTP.
- **Potential federation**: Apollo Federation entity resolvers receive a batch of `representations` and resolve them all at once — structurally identical to a DataLoader batch function. Adopting DataLoader now means federation entity resolvers are already written if ADR-013 is revisited.

## Options

**Option 1 — `graph-gophers/dataloader` v7**
Generics-based, battle-tested. Per-request loader instances coalesce `.Load(key)` calls within a tick into one batch. Batch function wraps `Keys` type; result wraps `*Result[V]` — more ceremony than necessary.

**Option 2 — `vikstrous/dataloadgen`**
Pure library (not a generator despite the name). Batch function signature is `func(ctx, []K) ([]V, []error)` — idiomatic Go, maps directly onto existing repository method signatures with no adapter layer. ~200 lines, no transitive dependencies. Type-safe via generics.

## Decision

**Adopt `vikstrous/dataloadgen`** (Option 2).

### The convention

> **Resolvers never call single-item repository methods by ID. Every `GetX(ctx, id)` call inside a resolver is a loader call.**

The loader inventory grows naturally with the schema. New entity types get a loader when the first resolver that needs them is written.

### Pattern

```
AroundOperations middleware
  └─ construct Loaders{} (one field per entity type)
  └─ inject into ctx

resolver
  └─ loaders := LoadersFromCtx(ctx)
  └─ entity, err := loaders.PlayerByPlayerSeasonID.Load(ctx, pm.PlayerSeasonID)

batch function
  └─ signature: func(ctx context.Context, ids []int) ([]*domain.Player, []error)
  └─ calls repo.FindByIDs(ctx, ids)   // one round-trip
  └─ returns slice positionally matched to input ids
```

The batch function signature aligns directly with repository `FindByIDs` methods — batch functions are one-liners in most cases.

### Loader placement

All loaders for a service live in a single `Loaders` struct in `internal/interface/graphql/loaders.go`. The struct is constructed in `cmd/main.go`'s `AroundOperations` closure and injected via a typed context key. Resolvers retrieve it with a `LoadersFromCtx(ctx)` helper.

Cross-service loaders (e.g. AFL stats from FFL context, HTTP-backed) follow the same interface and are added when that work is scoped.

## Rationale

- **Generic**: the convention eliminates the N+1 class of problem rather than patching known instances
- **`dataloadgen` batch signature**: `func(ctx, []K) ([]V, []error)` is already the shape of repository `FindByIDs` methods — near-zero adapter code
- **Fixes all forms**: coalescing within a tick handles parent-driven, sibling-driven, and (via HTTP-backed loaders) cross-service N+1 uniformly
- **Federation-ready**: batch functions are the natural implementation of Apollo Federation entity resolvers
- **Thin**: ~200 lines, easy to reason about and replace

## Consequences

- **New dependency**: `github.com/vikstrous/dataloadgen` added to AFL and FFL `go.mod`
- **`loaders.go`** added to `internal/interface/graphql/` in each service — owns the `Loaders` struct, context key, and `LoadersFromCtx` helper
- **`AroundOperations` middleware** extended in both `cmd/main.go` to construct and inject loaders per request
- **Resolvers unchanged structurally**: still receive `ctx context.Context`; loaders are retrieved from context
- **Repository interfaces unchanged**: batch functions call existing `FindByIDs` methods; domain and application layers are unaffected
- **ADR-013 relationship**: does not change the no-federation decision; makes federation entity resolvers trivially cheap to implement if that decision is revisited
