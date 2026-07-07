---
status: accepted
date: 2026-07-07
scope: frontend
enforceable: true
rules:
  - "pure frontend logic (algorithms, stat calculations, formatters) gets vitest unit tests colocated as *.test.ts next to the source"
  - "vitest tests do not mount components and do not hit the network — component behaviour stays in Playwright e2e"
  - "npm run test:unit must pass before committing frontend logic changes"
---

# ADR-019: Vitest for Frontend Unit Tests

## Context

The frontend has accumulated pure algorithmic logic — the Hungarian assignment
solver behind Best Team (`utils/bestTeam.ts`), trend thresholds and the star
score formula (`utils/playerStats.ts`) — whose correctness cannot reasonably be
asserted through Playwright e2e. E2e can show that players moved; it cannot show
that the assignment was optimal or that a trend arrow fires at exactly the
threshold. Until now the frontend's only test runner was Playwright.

## Decision

Add **vitest** as a devDependency of `frontend/web` for unit-testing pure
TypeScript logic.

- Tests live next to the code they test: `src/**/foo.test.ts`.
- Scope is pure functions and modules: algorithms, stat maths, formatters,
  data mapping. No component mounting, no DOM, no network.
- Component and interaction behaviour remains the job of the Playwright e2e
  suite (`e2e/`), which keeps its isolation model (ADR-free, see
  `ai/architecture/testing.md`).
- `npm run test:unit` runs the suite; it is fast enough to run on every change.

## Rationale

- **Vitest over Jest:** shares Vite's config and transform pipeline (the
  project is already Vite-based, ADR-011), native ESM + TypeScript support,
  no extra Babel toolchain.
- **Colocated tests** keep the algorithm and its spec in one place, matching
  how the Go services colocate `_test.go` files.
- **Narrow scope** avoids a parallel component-testing stack; Playwright
  already covers user-visible behaviour end to end.

## Alternatives considered

- **Jest:** would introduce a second transform pipeline alongside Vite.
  Rejected.
- **Only e2e:** cannot assert algorithmic optimality or threshold edges;
  silent regressions in Best Team would ship. Rejected.
- **Node built-in test runner:** no TS/Vite integration; would need its own
  build step. Rejected.
