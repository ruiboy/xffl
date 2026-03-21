# Agent System Prompt

You are working in a SOA monorepo. All architectural rules are defined in `ai/architecture/principles.md`

## Before You Code

1. Read `ai/plans/current-sprint.md`
2. Read `ai/architecture/` — `repo-map.md`, `service-map.md`, `bounded-contexts.md`
3. Check `ai/decisions/decisions.md` for relevant decisions; read full ADRs only when you need detail

## Development Process

All non-trivial tasks MUST begin by creating or resetting `ai-runtime/current-task.md`. Do not proceed to implementation until the Understand and Test Plan sections are filled.

1. **Understand** — identify affected services, relevant ADRs, bounded contexts
   → Record in current-task.md: Summary, affected services, relevant ADRs
2. **Test plan** — define what tests are needed before writing code
   → Record in current-task.md: test cases under Steps
3. **Implement** — write failing tests, then minimal implementation
   → Record in current-task.md: files changed, decisions made
4. **Validate** — run tests, run /checkarch and /checkdoc
   → Record in current-task.md: test results, validation results
5. **Reflect** — note assumptions, risks, and anything learned
   → Record in current-task.md: Reflection section
   → If a systemic issue emerged, propose an update to principles, ADRs, or prompts