# Current Sprint — Phase 21: UX Navigation

**Sprint goal:** Consistent, round-aware navigation across AFL and FFL — header links land on the live round, cross-domain switching is a first-class affordance, and DataOps is always reachable.

See `plans/ideas.md` for full descriptions of each item.

---

## Tasks

- [x] NAV-1: Update header AFL/FFL links to navigate to live round (`/afl/rounds/:liveRoundId`, `/ffl/rounds/:liveRoundId`)
- [x] NAV-2: Cross-domain round-aware navigation — header AFL/FFL/DataOps links follow a session-scoped "selected round" (defaults to live round, sticks to whichever round you're viewing, syncs the corresponding round across domains); replaces the originally-planned cross-domain pill, which was built then superseded by this approach
- [x] NAV-3: Add Ladder pill as first item in RoundNav (ladder icon + hover tooltip)
- [x] NAV-4: Add DataOps icon link to header right-side nav (always visible; links to live round)
- [x] NAV-5: Replace DataOps round dropdown with RoundNav component
- [x] PAGE-7: Enhance current-round pill with pulsing indicator when `start_dt` = today
- [x] Playwright tests for new navigation elements
