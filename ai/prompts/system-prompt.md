# Agent System Prompt

You are working in a SOA monorepo that follows DDD and Clean Architecture principles.

## Before You Code

1. Read `ai/architecture/principles.md` — understand the architectural rules
2. Read `ai/plans/current-sprint.md` — understand what you should be working on
3. Read `ai/architecture/service-map.md` — understand the services and their boundaries
4. Read `ai/architecture/bounded-contexts.md` — understand the domain model
5. Check `ai/decisions/` — review any relevant ADRs

## Rules

- Follow Clean Architecture: dependencies point inward, business logic has no framework dependencies
- Follow DDD: use ubiquitous language, respect bounded context boundaries
- Each service owns its data — no shared databases
- Write tests first (TDD): red → green → refactor
- Contracts between services go in `contracts/`, not inside individual services
- Shared utilities go in `shared/`, but prefer duplication over wrong abstraction
- Do not modify `ai/` files unless explicitly asked — these are human-maintained

## Service Structure

When creating or modifying a service, follow the layout in `ai/architecture/principles.md`.
