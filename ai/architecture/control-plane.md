# AI Control Plane

The ai/ directory is a declarative control plane for AI agents. Human architects define what to build and why; agents handle how.

## Philosophy

- **Human-in-the-loop** — human is architect (what/why), agent is implementer (how)
- **Structured environment, not autonomous agent** — no agent loops, no step automation
- **Agent-agnostic** — any LLM tool (Claude Code, Cursor, Aider) reads ai/ and ai-runtime/
- **Observable** — working memory in ai-runtime/ makes agent reasoning visible
- **Incremental** — evolve based on real usage, not hypothetical needs

## Layers

| Layer | Directory | Purpose | Who writes |
|-------|-----------|---------|------------|
| Control Plane | `ai/` | Architecture, decisions, plans, prompts | Human (agents read-only) |
| Runtime | `ai-runtime/` | Working memory, execution state | Agent (ephemeral, gitignored) |
| Agent Config | `.claude/` | Skills, hooks, settings | Human (Claude Code only) |

## Constraints

- `ai/` is read-only for agents — human maintains architectural intent
- `ai-runtime/` is gitignored — ephemeral, per-session state
- `.claude/` is vendor-specific — enhances but is not required
- Architecture and ADRs override code (see principles.md)

## What this is not

- Not an agent framework — no autonomous planning or task loops
- Not vendor-locked — the core system works without .claude/
- Not speculative — features are added when needed, not before
