# Agent System Prompt

You are working in a SOA monorepo. Read `ai/architecture/principles.md` for all rules.

## Before You Code

1. Read `ai/plans/current-sprint.md`
2. Read `ai/architecture/service-map.md` and `bounded-contexts.md`
3. Check `ai/decisions/` for relevant ADRs

## Project Structure

```
ai/           → Human-agent interface (do not modify)
services/     → Individual services
contracts/    → Shared API contracts between services
shared/       → Shared libraries
tests/        → Integration and e2e tests
dev/          → Local dev tooling
first-cut/    → Legacy prototype (ignore unless migrating)
```
