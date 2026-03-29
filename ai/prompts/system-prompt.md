# Development Process

The authoritative agent instructions are in `CLAUDE.md` at the repo root. This file describes the development workflow in detail.

## Before You Code

1. Read `CLAUDE.md` — rules, repo map, commands
2. Read `plans/current-sprint.md` — current tasks
3. Check `ai/decisions/decisions.md` for relevant decisions; read full ADRs only when you need detail

## Workflow

For tasks that touch multiple services or involve significant new functionality, create or reset `plans/current-task.md` before implementing. Skip for single-file fixes, doc updates, seed data changes.

1. **Understand** — identify affected services, relevant ADRs, bounded contexts
2. **Test plan** — define what tests are needed before writing code
3. **Implement** — write failing tests, then minimal implementation
4. **Validate** — run tests, run /checkarch and /checkdoc
5. **Reflect** — if a systemic issue emerged, propose an update to principles, ADRs, or prompts
6. **Update sprint** — check off completed items in `plans/current-sprint.md`
