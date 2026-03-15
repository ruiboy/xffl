# Architecture Principles

**Language:** Go | **Testing:** TDD (red → green → refactor)

## Clean Architecture Layers (inside → out)

1. **Domain** — entities, value objects, domain events, repository interfaces
2. **Application** — use cases, DTOs, port interfaces
3. **Infrastructure** — DB, messaging, external APIs, implement ports
4. **Interface** — HTTP/gRPC handlers

Dependencies point inward; never outward. Business logic has zero framework dependencies.

## DDD

- Each service = one bounded context
- Ubiquitous language throughout
- Aggregates enforce consistency; domain events cross context boundaries
- Repositories abstract persistence

## SOA

- Services are independently deployable, organized around business capabilities
- Each service owns its data — no shared database schemas
- Services communicate through contracts in `contracts/`
- Prefer async; sync only when necessary

## GraphQL

- Every query starts from a domain root — an aggregate the user naturally thinks in (e.g., Season, Club).
- Related data is accessed by traversing edges, not by separate top-level queries.
- Internal join entities (ClubMatch, PlayerSeason) are not query roots — they appear as nested fields of their parent.

## Rules

- Prefer duplication over wrong abstraction in `shared/`
- Domain logic must be tested
- Prefer TDD
- Prefer small incremental commits
- Ask questions when requirements are unclear
- Do not modify `ai/` files unless explicitly asked

## Service Layout

```
services/<name>/
├── cmd/
├── internal/
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   └── interface/
└── go.mod
```
