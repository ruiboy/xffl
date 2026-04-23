# Current Sprint — Phase 17: UX Improvements

**Sprint goal:** Iterative frontend improvements driven by TBD changes — work continues until done. No fixed scope; tasks are added and checked off as the session progresses.

## In Progress — Team Builder

- [x] Show player scores alongside names (starters and bench)
- [x] Position group totals in parentheses next to group heading
- [x] Grand total in Team summary bar; starters/bench count muted and before total
- [x] Status badge per player (right-aligned, before score); score suppressed unless status = Played
- [x] Scoring formula for multiplier positions (Goals ×5, Marks ×2, Tackles ×4) — `utils/scoring.ts` shared util
- [ ] TBD further Team Builder improvements

## Backlog — Phase 17

- [ ] Player pages — career stats, season history, club timeline
- [ ] Team pages — squad, round-by-round scores, season summary
- [ ] Richer stat data surfaced in existing views
- [ ] Performance: break up monolithic GraphQL queries; DataLoader pattern for N+1 elimination

## Ideas

- **Drag-and-drop in Team Builder manage mode** — feasible with `vue-draggable-plus` (SortableJS wrapper). Main complexity is enforcing per-position slot limits via `onMove` callbacks. Deferred — current button UI is functional; revisit once data flows settle.