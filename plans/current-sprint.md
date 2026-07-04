# Current Sprint — Phase 23: UX — Player Intelligence

**Sprint goal:** Richer player and club context for FFL decision-making.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [x] PAGE-1: AFL player season view (`/ffl/afl/player-seasons/:id`) — MEDIAN backend, status pills, season stats table, row-level match links, header breadcrumb links, stat label fixes
- [x] PAGE-3: AFL club season view (`/ffl/afl/club-seasons/:id`) — all players at an AFL club with FFL ownership, stats, and * avg; entry from AFL ladder club names
- [x] Club Match page (`/ffl/club-matches/:id`) — read-only view of a club's team and scores; linked from FFL match view; Team Builder link shown when club is selected
- [x] Frontend architecture: formalised three-namespace page hierarchy (FFL / AFL / AFL Lens) in `frontend.md`; consistent page naming and route conventions across `router.ts`, `frontend.md`, and `ideas.md`
