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

- Every query must start from a domain root (an aggregate the user naturally thinks in, e.g., Season or Club).
- Related data is accessed by traversing edges, not by separate top-level queries.
- Internal join entities (ClubMatch, PlayerSeason) are not query roots — they appear as nested fields of their parent.

## Testing

- **Domain** — unit tests. Pure logic, no mocks, no infrastructure.
- **Interface (GraphQL)** — integration tests against the running handler with real dependencies (DB, messaging).
- **Frontend** — end-to-end Playwright tests for all user-facing features.
- Do not test generated code (gqlgen models, generated resolvers).
- Table-driven tests for any function with more than one interesting input.

## Rules

### Architecture Authority
- Architecture and ADRs override the codebase.
- If code conflicts with architecture or ADRs, propose a fix.

### Boundaries
- Do not introduce new dependencies, services, or infrastructure without an ADR.
- Services must remain isolated; do not import code from another service.
- Prefer duplication over incorrect abstractions in `shared/`.

### Development Workflow
- Prefer the simplest solution that satisfies current requirements.
- Prefer TDD.
- Prefer small, incremental commits.

### Agent Behaviour
- **Never commit unless explicitly asked.** The user will say "commit", "commit that", "commit and push", etc. Do not auto-commit after completing work — the user needs to review diffs first.
- When requirements are unclear: Ask a question, propose possible options, wait for confirmation before implementing.
- Do not modify `ai/` files unless explicitly instructed.
- Agents may check off items in `plans/current-sprint.md` and `plans/current-task.md` as work is completed. Material changes to roadmap or sprint scope require discussion with the user.

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
