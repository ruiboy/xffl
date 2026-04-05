# Frontend Architecture

## Overview

Single Vue 3 SPA (`frontend/web/`) connecting to the gateway at port 8090 via Apollo Client.

**FFL is the app's front door** — the root route (`/`) is the FFL home. AFL lives under `/afl`.
The primary audience is FFL club managers who use the app to track fantasy scores and build teams.

## Page Hierarchy

### FFL (primary)

| Page | Route |
|------|-------|
| Home | `/` |
| Round | `/ffl/seasons/:seasonId/rounds/:roundId` |
| Match | `/ffl/seasons/:seasonId/matches/:matchId` |
| Team Builder | `/ffl/seasons/:seasonId/rounds/:roundId/team-builder` |
| Players (admin) | `/ffl/players` |

**Money-shot views:** Match (head-to-head fantasy scores in real time) and Team Builder (weekly team selection).

### AFL (supporting)

| Page | Route |
|------|-------|
| Home | `/afl` |
| Round | `/afl/seasons/:seasonId/rounds/:roundId` |
| Match | `/afl/seasons/:seasonId/matches/:matchId` |
| Admin Match | `/admin/afl/seasons/:seasonId/matches/:matchId` |

AFL pages exist to enter real-world match stats, which feed into FFL scoring.

## Key Design Decisions

- **FFL front door** — FFL managers are the primary users; AFL is a data-entry tool accessed via navbar.
- **Apollo routing** — the Apollo Client uses regex to route FFL queries to `/ffl/graphql` and AFL queries to `/afl/graphql` on the gateway. This is acknowledged as fragile (see `plans/revisit.md`).
- **Club logos** — AFL logos at `public/images/clubs/`, FFL logos at `public/images/ffl-clubs/`. Each feature has a `utils/clubLogos.ts` that maps club names to file paths.
- **No cross-feature imports** — `features/afl/` and `features/ffl/` are independent; shared UI lives in `components/`.

## Source Layout

```
frontend/web/src/
  features/
    afl/
      api/          — GraphQL queries + mutations
      components/   — MatchSummary, LadderTable, PlayerStatsTable, TopPlayers, RoundNav
      utils/        — clubLogos.ts
      views/        — HomeView, RoundView, MatchView, AdminMatchView
    ffl/
      api/          — GraphQL queries + mutations
      components/   — MatchSummary, LadderTable, SquadTable, RoundNav, StatusBadge
      utils/        — clubLogos.ts
      views/        — HomeView, RoundView, MatchView, TeamBuilderView, PlayersView
  components/       — NavBar (shared)
  router/           — Vue Router config
```
