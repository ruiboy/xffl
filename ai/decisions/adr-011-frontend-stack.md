---
status: accepted
date: 2026-03-23
scope: frontend
enforceable: true
---

# ADR-011: Frontend Stack

## Context

Need a web frontend for managing fantasy teams, viewing AFL stats, searching, and displaying the ladder. Want a custom visual identity — not boxed into a component library theme — while still getting accessible, battle-tested UI behaviour out of the box.

## Decision

### Pass 1 — Core (scaffold + first view)

| Layer | Choice |
|-------|--------|
| Framework / Build | Vue 3 + TypeScript + Vite |
| Server State / GraphQL | Apollo Client |
| Styling | Tailwind CSS |

### Pass 2 — When the first view needs them

| Layer | Choice |
|-------|--------|
| UI Components | PrimeVue unstyled |
| State Management | Pinia |
| Hooks / Utilities | VueUse |

### Styling approach

- PrimeVue unstyled provides behaviour (tables, modals, dropdowns); Tailwind provides all visual styling.
- Small variant wrappers (class maps in `ui/variants.ts`) for consistent button/input/badge patterns — no component abstraction layer until pain is felt.

## Rationale

- **Vue 3 + Vite:** Lightweight, fast HMR, standard pairing.
- **Apollo Client:** Purpose-built for GraphQL; handles caching, queries, mutations against the gateway.
- **Tailwind CSS:** Full control over appearance without fighting a theme system.
- **PrimeVue unstyled:** Accessible component behaviour without imposed styles. Added in pass 2 so each dependency earns its place.
- **Pinia:** Clean separation — Pinia for local UI state, Apollo cache for server state. Added when UI state management is actually needed.
- **VueUse:** Zero-cost reactive utilities. Added when hand-rolling becomes tedious.

## Alternatives considered

- **PrimeVue themed:** Faster to start but locks into a visual identity. Rejected — want a custom look.
- **Headless UI (Radix Vue):** Good primitives but smaller ecosystem than PrimeVue. Could supplement later for very custom components.
- **No component library:** Maximum control but re-implementing accessibility (focus traps, ARIA, keyboard nav) is not worth it for a hobby project.