# FFL Frontend — Page Inventory

**Primary audience:** FFL team managers (club owners)

## Pages

### 1. Home (`/`)

FFL becomes the app's front door. AFL moves to `/afl`.

- FFL ladder for current season
- Current round's matches with fantasy scores
- Round navigation
- Link to AFL section

### 2. Round (`/ffl/seasons/:seasonId/rounds/:roundId`)

- All matches in the round with scores
- Top fantasy scorers across the round
- Round navigation

### 3. Match (`/ffl/seasons/:seasonId/matches/:matchId`) — money shot

Head-to-head match detail, the view managers watch as scores roll in.

- Two club rosters side by side
- Each player: name, FFL position, status (played/DNP/named), fantasy score
- Bench/sub/interchange indicators (did a sub fire? did an interchange swap?)
- Club fantasy score totals

### 4. Team Builder (`/ffl/seasons/:seasonId/rounds/:roundId/team-builder`) — money shot

Where managers build their 22-player lineup each week from their 30-player roster.

- View roster (30 players) with AFL stats from recent rounds
- Assign players to FFL positions (goals, kicks, handballs, marks, tackles, hitouts, star)
- Designate starters (22) vs bench (8) with backup/interchange positions
- Compare different lineup arrangements
- Submit lineup

**v1 (this sprint):** Stub UI with layout, position slots, roster display. No persistence.
**v2 (backend wiring):** Add `aflPlayerId` to FFL Player, `setFFLLineup` mutation, roster query. Wire UI to real data.

### 5. Players (`/ffl/players`)

Admin/setup view for player management.

- CRUD: create, edit, delete FFL players
- Assign/remove players to/from club seasons (roster management)

### 6. AFL Section (existing, re-routed)

- AFL Home → `/afl`
- AFL Round → `/afl/seasons/:seasonId/rounds/:roundId`
- AFL Match → `/afl/seasons/:seasonId/matches/:matchId`
- AFL Admin Match → `/admin/afl/seasons/:seasonId/matches/:matchId`

## Backend changes needed (end of sprint)

1. **`aflPlayerId` on FFL Player** — domain + schema + migration to link FFL players to AFL players
2. **`setFFLLineup` mutation** — upsert batch of PlayerMatch entries for a club match
3. **Roster query** — expose club's PlayerSeason entries via GraphQL (with linked AFL player data)
