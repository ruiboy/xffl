# Architecture Principles

**Language:** Go | **Testing:** TDD (red → green → refactor)

## Clean Architecture Layers (inside → out)

1. **Domain** — entities, value objects, domain events, repository interfaces
2. **Application** — use cases, DTOs, port interfaces
3. **Infrastructure** — DB, messaging, external APIs
4. **Interface** — HTTP/gRPC handlers

Dependencies point inward. Business logic has zero framework dependencies.

## DDD

- Each service = one bounded context
- Ubiquitous language throughout
- Aggregates enforce consistency; domain events cross context boundaries
- Repositories abstract persistence

## SOA

- Services are independently deployable, organized around business capabilities
- Each service owns its data — no shared databases
- Services communicate through contracts in `contracts/`
- Prefer async; sync only when necessary

## Rules

- Prefer duplication over wrong abstraction in `shared/`
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
