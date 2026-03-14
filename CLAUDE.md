# CLAUDE.md

## Before You Do Anything

Read these files in order:
1. `ai/prompts/system-prompt.md` — your operating instructions
2. `ai/plans/current-sprint.md` — what to work on
3. `ai/architecture/principles.md` — how to build it

## Project Structure

```
ai/           → Human-agent interface (read before coding, do not modify)
services/     → Individual services (DDD + Clean Architecture)
contracts/    → Shared API contracts between services
shared/       → Shared libraries
tests/        → Integration and e2e tests
dev/          → Local dev tooling (Docker Compose, scripts)
first-cut/    → Legacy prototype (ignore unless migrating specific code)
```

## Key Conventions

- **Language:** Go
- **Architecture:** SOA + DDD + Clean Architecture
- **Testing:** TDD (write tests first)
- **Services:** Each service is self-contained with its own `go.mod`
- **Dependencies flow inward:** Domain ← Application ← Infrastructure ← Interface
