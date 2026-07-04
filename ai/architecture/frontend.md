# Frontend Architecture

## Overview

Single Vue 3 SPA (`frontend/web/`) connecting to the gateway at port 8090 via Apollo Client.

**FFL is the app's front door** — the root route (`/`) is the FFL home. AFL lives under `/afl`.
The primary audience is FFL club managers who use the app to track fantasy scores and build teams.

## Page Hierarchy

There are three namespaces, each with a distinct purpose:

### FFL

Pure FFL-domain views. Primary audience: FFL club managers.

| Page | Route |
|------|-------|
| Home | `/ffl` |
| Round | `/ffl/rounds/:roundId` |
| Match | `/ffl/matches/:matchId` |
| Club Season | `/ffl/club-seasons/:clubSeasonId` |
| Club Match | `/ffl/club-matches/:clubMatchId` |
| Team Builder | `/ffl/club-matches/:clubMatchId/edit` — edit-mode overlay on Club Match; only for the selected club |
| Data Ops | `/ffl/data-ops` |

**Money-shot views:** Match (head-to-head fantasy scores in real time) and Team Builder (weekly team selection).

### AFL

Pure AFL-domain views. Used for real-world stat entry, which feeds FFL scoring.

| Page | Route |
|------|-------|
| Home | `/afl` |
| Round | `/afl/rounds/:roundId` |
| Match | `/afl/matches/:matchId` |
| Match Edit | `/afl/matches/:matchId/edit` |

### AFL Lens (`/ffl/afl/`)

FFL's view of AFL entities — the bridge between AFL data and FFL decision-making. No pure AFL player or club pages exist; these are the FFL-intelligence equivalents.

| Page | Route |
|------|-------|
| Player Season | `/ffl/afl/player-seasons/:aflPlayerSeasonId` |
| Club Season | `/ffl/afl/club-seasons/:clubSeasonId` |

## Key Design Decisions

- **FFL front door** — FFL managers are the primary users; AFL is a data-entry tool accessed via navbar.
- **Apollo routing** — the Apollo Client sends all queries to a single `/query` endpoint on the gateway. Apollo Router (Federation) handles cross-service composition.
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
      views/        — HomeView, RoundView, MatchView, AdminMatchView, ClubSeasonView
    ffl/
      api/          — GraphQL queries + mutations
      components/   — MatchSummary, LadderTable, SquadTable, RoundNav, StatusBadge
      utils/        — clubLogos.ts
      views/        — HomeView, RoundView, MatchView, SquadView, TeamBuilderView, AFLPlayerSeasonView
  components/       — NavBar (shared)
  app/              — router.ts
```
