---
status: accepted
date: 2026-04-27
scope: infrastructure
enforceable: true
rules:
  - "all synchronous cross-service RPC uses Twirp — no ad-hoc HTTP handlers between services"
  - "proto definitions live in contracts/proto/; generated Go code in contracts/gen/"
  - "generated Twirp code is infrastructure — never imported in domain or application layers"
  - "application layer depends on a port interface, not the generated client"
  - "buf is the code generation toolchain (buf.gen.yaml at repo root)"
  - "Twirp servers are mounted on the same HTTP port as GraphQL under the /twirp/ path prefix"
---

# ADR-018: Twirp for Synchronous Cross-Service RPC

## Context

ADR-004 established async events (PG LISTEN/NOTIFY) for cross-service communication where the producer does not need to wait for the consumer — e.g. `AFL.PlayerMatchUpdated` triggering FFL score recalculation.

Phase 18 import use cases introduce a different communication pattern: **synchronous request-response** where the calling service cannot proceed without the answer. The first case is FFL resolving AFL player records during squad and round-team imports. FFL needs an `afl_player_id` before it can write `ffl.player` records; there is no event-driven alternative — the import is a blocking operation.

FFL cannot read AFL's database schema directly (ADR-003), and the existing GraphQL endpoints are user-facing APIs not designed for internal service-to-service calls.

## Options

**Option 1 — Twirp**
RPC over HTTP/1.1; proto-defined contracts; `buf` generates type-safe client and server stubs from `.proto` files. The proto file is the contract; codegen enforces it on both sides. No HTTP/2 requirement.

**Option 2 — Plain HTTP/JSON**
Hand-written handlers and client structs. No build toolchain additions; easier to start. The contract between client and server is implicit — nothing enforces they stay in sync across changes.

**Option 3 — gRPC**
Full HTTP/2, bidirectional streaming, larger ecosystem. No streaming needed here; HTTP/2 adds infra complexity for no gain in this context.

**Option 4 — Shared Go package in `shared/`**
Move player lookup logic to `shared/` so both services can call it. Violates ADR-005 (service isolation) — FFL would gain access to AFL domain logic. Prefer duplication over incorrect abstractions in `shared/` (principles).

## Decision

**Adopt Twirp (Option 1).**

### When to use Twirp vs events

| Situation | Mechanism |
|-----------|-----------|
| Caller must receive a response before proceeding | Twirp |
| Producer notifies; consumer acts independently | Events (ADR-004) |

Events remain the default for cross-service communication. Twirp is used only when the calling service genuinely blocks on the response.

### Contract location

Proto definitions live in `contracts/proto/<service>/v1/`. Generated Go code is written to `contracts/gen/` as a separate Go module, imported by both the server service and the client service.

```
contracts/
├── events/          # existing — PG event schemas
│   ├── events.go
│   └── go.mod
├── proto/           # Twirp source definitions
│   └── afl/v1/
│       └── player_lookup.proto
└── gen/             # buf-generated Go stubs (do not edit)
    ├── go.mod
    └── afl/v1/
        └── player_lookup.pb.go
        └── player_lookup_twirp.go
```

`buf.gen.yaml` lives at the repo root. `contracts/gen/` is added to `go.work`.

### Port interface pattern

The application layer defines a port interface; the Twirp client adapter implements it. Use cases never import generated stubs.

```go
// services/ffl/internal/application/ports.go
type PlayerLookup interface {
    FindCandidates(ctx context.Context, club string, name string) ([]PlayerCandidate, error)
}

// services/ffl/internal/infrastructure/rpc/afl_player_lookup.go
// implements PlayerLookup using the generated Twirp client
```

### Server mounting

AFL mounts a Twirp server alongside its existing GraphQL handler. Both share the same HTTP port; Twirp routes are prefixed `/twirp/` (Twirp default). No new port or listener is required.

### Toolchain

`buf` manages proto linting, breaking-change detection, and codegen. The `buf.gen.yaml` at the repo root targets `contracts/proto/` and outputs to `contracts/gen/`.

A `just proto-gen` recipe runs `buf generate`. CI lint runs `buf lint` + `buf breaking` against the last known good state.

## Rationale

- **Contract enforcement** — the proto file is the single source of truth; `buf generate` keeps client and server stubs in sync automatically. Plain HTTP/JSON has no equivalent guarantee.
- **HTTP/1.1** — no HTTP/2 requirement; works with the existing reverse proxy and `net/http` stack unchanged.
- **Simpler than gRPC** — no streaming, no HTTP/2, friendlier error model (maps to HTTP status codes).
- **Port interface pattern** — generated code stays in infrastructure; application and domain layers are unaffected. Swapping Twirp for another transport later is one adapter change.
- **`contracts/` cohesion** — proto definitions sit alongside event schemas; all inter-service contracts are in one place.

## Consequences

- **New toolchain dependency**: `buf` added to dev tooling (not a runtime dependency). `just proto-gen` recipe added to `justfile`.
- **New Go module**: `contracts/gen/` — added to `go.work`; imported by AFL (server) and FFL (client).
- **AFL service** — gains a Twirp server handler mounted at `/twirp/`; no change to GraphQL handler.
- **FFL service** — gains an `infrastructure/rpc/` adapter implementing the `PlayerLookup` port.
- **Application and domain layers unaffected** — use cases depend on the port interface only.
- **ADR-004 unchanged** — async events remain the default; Twirp is additive for synchronous cases only.
- **ADR-013 unchanged** — Twirp is internal infrastructure, not a user-facing API; no impact on the no-federation decision.
