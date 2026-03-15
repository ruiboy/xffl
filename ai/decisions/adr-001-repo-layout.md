# ADR-001: AI-Optimized Repository Layout

**Status:** Accepted
**Date:** 2026-03-14

## Context

We are building a SOA system with multiple services using DDD and Clean Architecture. Development will be driven by AI agents (Claude Code) working alongside human architects. We need a repo layout that clearly separates human architectural decisions from AI implementation work.

## Decision

Adopt an AI-optimized monorepo layout with an `ai/` directory that serves as the interface between humans and agents.

```
repo/
├── ai/           # Human → Agent interface (architecture, plans, decisions, prompts)
├── services/     # Individual services (each follows Clean Architecture)
├── contracts/    # Shared API contracts between services
├── shared/       # Shared libraries and utilities
├── tests/        # Integration and end-to-end tests
└── dev/          # Local development tooling (Docker Compose, scripts)
```

**Key principle:** Agents read `ai/` before coding. Humans maintain `ai/` to steer agent work.

## Consequences

- Clear separation of concerns between architecture and implementation
- Agents always have up-to-date context on principles, plans, and decisions
- ADRs provide traceable decision history
- Roadmap-driven development keeps agents focused on current priorities
