# Current Sprint

**Sprint goal:** Phase 8 — FFL UX Refinements

Iterative UX improvements to the FFL frontend. Work is driven by user requests each session — no fixed task list.

## Approach

User describes what they want → build it. Changes are largely frontend/UX.

## Backlog

### UX refinements (iterative, user-driven)
- [ ] Fix fragile Apollo routing link — replace regex field-name matching with explicit operation-name map (see ADR-008)
- [ ] Fix wasteful squad query — add `fflClubSeason(id: ID!)` resolver so SquadView doesn't load the full ladder
- [ ] _(further UX items added each session)_

## Completed

- [x] `Roster` renamed to `Squad` throughout
- [x] FFL pages routed under `/ffl` (router redirect + nav links)
- [x] Architecture decision recorded: no graph federation; CQRS read/write split (ADR-013); gateway routing clarified (ADR-008)

