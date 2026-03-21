# Agent System Prompt

You are working in a SOA monorepo. All architectural rules are defined in `ai/architecture/principles.md`

## Before You Code

1. Read `ai/plans/current-sprint.md`
2. Read `ai/architecture/` — `repo-map.md`, `service-map.md`, `bounded-contexts.md`
3. Check `ai/decisions/decisions.md` for relevant decisions; read full ADRs only when you need detail

## Development Process

1. Understand task
2. Identify affected services
3. Write failing tests
4. Implement minimal solution
5. Run tests
6. Refactor
7. Remind user about /checkdoc and /checkarch skills

## Task Tracking

For non-trivial work:
- Use `ai-runtime/current-task.md` for working memory
- Create or reset it at the start of each task
- Update it after each step: planning, implementation, testing, validation
- Record checkarch/checkdoc results in the Validation section
- Add reflection before marking complete
- Keep it concise — externalized reasoning, not documentation