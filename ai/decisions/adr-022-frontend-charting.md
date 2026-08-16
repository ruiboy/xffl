---
status: accepted
date: 2026-08-15
scope: frontend
enforceable: true
rules:
  - "chart.js (core) + vue-chartjs is the standard for frontend data visualization"
  - "register only the Chart.js elements/scales/plugins a chart actually uses (e.g. `ChartJS.register(BarElement, CategoryScale, LinearScale, Tooltip)`) — never import `chart.js/auto`"
  - "adopting a different charting library requires a follow-up ADR"
---

# ADR-022: Chart.js for Frontend Charts

## Context

The FFL club-season page needs a round-by-round score graph (bars colored by
win/loss/bye), with a follow-up phase planned to overlay ladder position as a
line on the same chart (dual-axis combo chart). The frontend has no charting
library today (`frontend/web/package.json` has no chart dependency).

## Decision

Add **chart.js** and **vue-chartjs** (the Vue 3 wrapper) as dependencies of
`frontend/web`.

- Register only the specific Chart.js building blocks a given chart uses
  (`ChartJS.register(BarElement, CategoryScale, LinearScale, Tooltip, ...)`),
  not the `chart.js/auto` bundle, to keep the bundle tree-shaken.
- Chart color/theme choices follow the existing dark/light convention driven
  by `useTheme()` (`src/composables/useTheme.ts`), the same composable
  `LadderTable.vue` already uses for its heatmap coloring.

## Rationale

- **Combo charts:** Chart.js natively supports a bar dataset and a line
  dataset sharing one chart with independent (dual) axes — exactly the shape
  needed once ladder position is overlaid on round scores. Avoids rebuilding
  the chart on a different library for that follow-up.
- **Per-mark styling:** per-bar and per-point color functions make win/loss/
  bye coloring straightforward.
- **Lightweight:** canvas-based, small footprint compared to ECharts, no
  build-step integration burden compared to D3.
- **Vue 3 fit:** `vue-chartjs` is a thin Vue 3 wrapper with no additional
  runtime dependencies beyond Chart.js itself.

## Alternatives considered

- **ECharts (`vue-echarts`):** more powerful (zoom, animated transitions,
  richer tooltips) but a heavier bundle and a steeper API for what's
  currently a fairly simple visual. Revisit if requirements grow into
  needing that interactivity.
- **D3:** full control, but building a combo bar+line chart from scratch is
  significantly more implementation effort for an equivalent result.
- **Hand-rolled SVG/CSS:** no new dependency at all, and sufficient for a
  single bar chart, but doesn't carry forward cleanly into the planned
  dual-axis ladder-position overlay. Rejected in favor of Chart.js given the
  follow-up phase is already planned.
