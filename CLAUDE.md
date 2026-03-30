# CLAUDE.md

AI agents must read this file before making changes to the repository.

## Repository Map

```
ai/              → Architecture, decisions, prompts (read-only for agents)
plans/           → Roadmap, current sprint, agent working memory
services/afl/    → AFL service (Go, GraphQL, port 8080)
services/ffl/    → FFL service (Go, GraphQL, port 8081)
services/gateway/→ Reverse proxy gateway (port 8090)
frontend/web/    → Vue 3 SPA (TypeScript, Vite, port 3000)
contracts/       → Shared event type definitions
shared/          → Shared Go packages (database, events)
dev/             → Docker Compose, seed data, dev tooling
first-cut/       → Legacy prototype (reference only)
```

## Non-Negotiable Rules

1. **Clean Architecture** — dependencies point inward (Domain ← Application ← Infrastructure ← Interface). Business logic has zero framework dependencies.
2. **Service isolation** — no cross-service imports, no shared DB schemas, communicate through contracts.
3. **`ai/` is read-only** — agents never modify files in `ai/` unless explicitly instructed.
3b. **`plans/` updates** — agents may check off sprint/task items as they complete work. Material changes to roadmap or sprint scope require discussion with the user.
4. **TDD** — write failing tests first, then minimal implementation.
5. **No new dependencies/services/infra without an ADR.**
6. **When unclear, ask** — propose options, wait for confirmation.

## Common Commands

```
just dev-up          # Start Postgres + Zinc (Docker)
just dev-down        # Stop infrastructure
just dev-reset       # Stop + delete all data
just dev-seed        # Load test data
just run-all         # Run AFL service + gateway + frontend
just run-afl         # AFL service only (port 8080)
just run-gateway     # Gateway only (port 8090)
just run-frontend    # Frontend only (port 3000)
just test-e2e        # Playwright e2e tests (requires run-all)
```

## Before Coding — Read These (tiered)

**Always before coding:**
- `plans/current-sprint.md` — what to work on now
- `ai/architecture/principles.md` — architecture rules and service layout

**Before architecture changes:**
- `ai/decisions/decisions.md` — ADR index with summary table
- `ai/architecture/domain.md`, `ai/architecture/service-map.md`

**For development workflow detail:**
- `ai/prompts/system-prompt.md` — development process (understand → test plan → implement → validate → reflect)

## Agent Working Memory

For non-trivial tasks, create or reset `plans/current-task.md` before implementing. Skip for single-file fixes, doc updates, seed data changes. This file is gitignored.

Template:
```markdown
# Current Task: <short title>

## Status
Not Started | In Progress | Done

## Summary
<what and why>

## Affected Services
- <service list>

## Steps
- [ ] step 1
- [ ] step 2

## Decisions Made
<updated as work progresses>
```
