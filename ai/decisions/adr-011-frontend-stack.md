---
status: accepted
date: 2026-03-14
scope: frontend
enforceable: false
---

# ADR-011: Frontend Stack

## Context

Need a web frontend for managing fantasy teams, viewing AFL stats, searching, and displaying the ladder.

## Decision

Vue 3 + TypeScript + Vite. Apollo Client for GraphQL. PrimeVue for UI components.

## Rationale

- **Why for Hobby:** Vue is lightweight and reactive, Vite gives instant HMR, Apollo is purpose-built for GraphQL, PrimeVue provides DataTable/Dialog/forms out of the box
- **Scale Path:** Add Pinia for state management, SSR with Nuxt, component library extraction
