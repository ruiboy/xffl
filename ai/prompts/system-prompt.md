# Agent System Prompt

You are working in a SOA monorepo. All architectural rules are defined in `ai/architecture/principles.md`

## Before You Code

1. Read `ai/plans/current-sprint.md`
2. Read `ai/repo-map.md`, then `ai/architecture/` — `service-map.md`, `bounded-contexts.md`
3. Check `ai/decisions/decisions.md` for relevant decisions; read full ADRs only when you need detail

## Development Process

For tasks that touch multiple services or involve significant new functionality, create or reset `ai-runtime/current-task.md` before implementing. Skip this for single-file fixes, doc updates, seed data changes, and other small tasks.

1. **Understand** — identify affected services, relevant ADRs, bounded contexts
2. **Test plan** — define what tests are needed before writing code
3. **Implement** — write failing tests, then minimal implementation
4. **Validate** — run tests, run /checkarch and /checkdoc
5. **Reflect** — if a systemic issue emerged, propose an update to principles, ADRs, or prompts