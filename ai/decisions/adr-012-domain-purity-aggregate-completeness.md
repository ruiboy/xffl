# ADR-012: Domain Purity and Aggregate Completeness

## Status
Accepted

## Context

Domain logic must remain deterministic, testable, and independent of infrastructure concerns.

A common failure mode in layered architectures is allowing domain entities to:
- Perform I/O (database, network, messaging)
- Lazily load dependent data
- Depend on repositories or infrastructure services

This introduces hidden side effects, breaks testability, and violates Clean Architecture principles.

Additionally, partially-loaded aggregates (missing required data) lead to unclear responsibilities:
- Domain logic cannot execute without additional data
- Data loading becomes implicit and scattered

## Decision

### Domain Purity

- Domain entities must be pure and side-effect-free.
- Domain code must not perform I/O or depend on infrastructure (DB, HTTP, messaging, etc).
- Domain logic must be executable entirely with in-memory data.

### Aggregate Completeness

- All data required for a domain operation must be provided before invoking domain methods.
- Domain methods must not fetch or request additional data.
- Do not pass repositories or infrastructure dependencies into domain entities.
- Aggregates contain child entities as embedded values, not as foreign key IDs. Related entities are traversed through the aggregate, not resolved separately.

### Aggregate Boundaries

- Aggregate boundaries are defined by consistency requirements, not database table structure.
- Each aggregate has a single root entity through which all access to child entities flows.

### Repository Responsibility

- Repositories must return aggregates in a valid, usable state for their intended operation.
- Do not return partially-loaded aggregates that require further data fetching.
- Repository methods may be use-case-specific (e.g. `GetMatchWithDetails`). Use-case-specific repository methods are preferred over generic methods that over-fetch.

### Time and Randomness

- Domain entities must not access system time or randomness directly.
- Time, IDs, and ordering decisions are provided by the application layer.

### Application Responsibility

- Application use cases orchestrate data loading via repositories.
- Use cases construct or retrieve fully-initialized aggregates.
- Use cases invoke domain logic and handle side effects (persistence, events).
- Use cases provide time, IDs, and any non-deterministic inputs to domain logic.

## Consequences

### Positive

- Domain logic is deterministic and easy to unit test.
- Clear separation of concerns between domain, application, and infrastructure.
- No hidden I/O or implicit dependencies.
- Aggregates enforce their invariants consistently.

### Negative

- Repository interfaces grow to support multiple use-case-specific queries. This is acceptable and by design — explicit data loading is preferred over implicit over-fetching.
- Some duplication in data loading logic may occur across use cases.
- Larger aggregate loads may impact performance if not carefully scoped.

## Enforceable Rules

- Domain packages must not import infrastructure, database, or network libraries.
- Domain entities must not accept repository or service interfaces as dependencies.
- Domain entities must not call `time.Now()`, `rand.*`, or UUID generators directly.
- All repository implementations must reside in the infrastructure layer.
- Application layer is responsible for coordinating repository calls before invoking domain logic.

## Notes

- Prefer explicit data loading over implicit or lazy loading.
- Prefer duplication over introducing leaky abstractions to "share" partially-loaded aggregates.