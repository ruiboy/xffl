# Architecture Principles

## Service-Oriented Architecture (SOA)

- System is composed of independently deployable services
- Services communicate through well-defined contracts
- Each service owns its data and persistence
- Services are organized around business capabilities, not technical layers
- Prefer asynchronous communication where possible; synchronous only when necessary

## Domain-Driven Design (DDD)

- Code models the business domain, not the database or UI
- Each service aligns to a bounded context
- Ubiquitous language: code uses the same terms as the domain experts
- Aggregates enforce consistency boundaries
- Domain events communicate state changes across contexts
- Entities have identity; Value Objects are defined by their attributes
- Repositories abstract persistence behind domain interfaces

## Clean Architecture

- Dependencies point inward: outer layers depend on inner layers, never the reverse
- Layers (inside → out):
  1. **Domain** — entities, value objects, domain events, repository interfaces
  2. **Application** — use cases / application services, DTOs, port interfaces
  3. **Infrastructure** — database, messaging, external APIs, framework code
  4. **Interface** — HTTP handlers, CLI, gRPC endpoints
- Business logic has zero dependency on frameworks, databases, or transport
- Use cases orchestrate domain objects; they don't contain business rules
- Interfaces are defined in inner layers, implemented in outer layers (Dependency Inversion)

## Service Structure Convention

Each service follows this internal layout:

```
services/<service-name>/
├── cmd/              # Entry points
├── internal/
│   ├── domain/       # Entities, value objects, domain events, repository interfaces
│   ├── application/  # Use cases, DTOs, port interfaces
│   ├── infrastructure/  # DB, messaging, external API implementations
│   └── interface/    # HTTP/gRPC handlers
├── go.mod
└── README.md
```
