# Current Sprint — Phase 21: UX Navigation

**Sprint goal:** Consistent, round-aware navigation across AFL and FFL — header links land on the live round, cross-domain switching is a first-class affordance, and DataOps is always reachable.

See `plans/ideas.md` for full descriptions of each item.

---

## Tasks

- [ ] NAV-1: Update header AFL/FFL links to navigate to live round (`/afl/rounds/:liveRoundId`, `/ffl/rounds/:liveRoundId`)
- [ ] NAV-2: Add cross-domain round pill to round view header — "↔ AFL/FFL Round N"; replaces bottom-of-page AFL Round link on FFL round view
- [ ] NAV-3: Add Ladder pill as first item in RoundNav (ladder icon + hover tooltip)
- [ ] NAV-4: Add DataOps icon link to header right-side nav (always visible; links to live round)
- [ ] NAV-5: Replace DataOps round dropdown with RoundNav component
- [ ] PAGE-7: Enhance current-round pill with pulsing indicator when `start_dt` = today
- [ ] Playwright tests for new navigation elements
