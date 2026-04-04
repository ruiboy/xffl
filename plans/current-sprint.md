# Current Sprint

**Sprint goal:** Phase 8 — FFL UX Refinements

Iterative UX improvements to the FFL frontend. Work is driven by user requests each session — no fixed task list.

## Approach

User describes what they want → build it. Changes are largely frontend/UX.

## Backlog

### UX refinements (iterative, user-driven)
- [x] Fix fragile Apollo routing link — replace regex field-name matching with explicit operation-name map (see ADR-008)
- [x] Fix wasteful squad query — add `fflClubSeason(seasonId, clubId)` resolver; rename `FFLSquadEntry` → `FFLPlayerSeason`; `squad` → `players` with connection pagination shape
- [x] Home/round page layout: circle round selector, ladder icon, inline headings, no matches on home
- [x] Round selector: filled active, ring for live round, ladder circle links home
- [x] FFL eagle logo in nav (hover scales 3×)
- [x] Settings cog dropdown with dark mode toggle (cookie-persisted)
- [x] Squad page: club name heading, search panel alongside player list
- [x] Team Builder: club name heading, Manage/Done pattern (Done saves lineup)
- [x] `FFLClubSeason.season` field added to GraphQL type and resolver
- [ ] _(further UX items added each session)_

 Completed

- [x] `Roster` renamed to `Squad` throughout
- [x] FFL pages routed under `/ffl` (router redirect + nav links)
- [x] Architecture decision recorded: no graph federation; CQRS read/write split (ADR-013); gateway routing clarified (ADR-008)

