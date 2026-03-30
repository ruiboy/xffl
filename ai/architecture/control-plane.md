# AI Control Plane

The ai/ directory is a declarative control plane for AI agents. Human architects define what to build and why; agents handle how.

## Philosophy

- **Human-in-the-loop** — human is architect (what/why), agent is implementer (how)
- **Structured environment, not autonomous agent** — no agent loops, no step automation
- **Agent-agnostic** — any LLM tool (Claude Code, Cursor, Aider) reads ai/ and plans/
- **Observable** — sprint checklist and task tracking make progress visible
- **Incremental** — evolve based on real usage, not hypothetical needs

## Layers

| Layer | Directory | Purpose | Who writes |
|-------|-----------|---------|------------|
| Control Plane | `ai/` | Architecture, decisions, prompts | Human (agents read-only) |
| Plans | `plans/` | Roadmap, sprint, working memory | Collaborative (agents check off items; scope changes need discussion) |
| Agent Config | `.claude/` | Skills, hooks, settings | Human (Claude Code only) |

## Source of Truth Hierarchy

When instructions conflict, architectural authority wins:

1. `ai/architecture/principles.md` + `ai/decisions/` — foundational rules and ADRs
2. `CLAUDE.md` — root agent instructions; references and delegates to the above
3. `ai/prompts/system-prompt.md` — workflow detail; must not contradict the above

CLAUDE.md is the entry point agents read first, but if it drifts from principles or ADRs, the architecture docs win.

## Constraints

- `ai/` is read-only for agents — human maintains architectural intent
- `.claude/` is vendor-specific — enhances but is not required
- Architecture and ADRs override code (see principles.md)

## What this is not

- Not an agent framework — no autonomous planning or task loops
- Not vendor-locked — the core system works without .claude/
- Not speculative — features are added when needed, not before
