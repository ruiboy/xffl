# Current Sprint

**Sprint goal:** Phase 6 — FFL Frontend

Build FFL views in the existing Vue 3 frontend. FFL becomes the app's main entry point. Primary audience is FFL team managers (club owners). Two money-shot views: Match (watching scores roll in) and Team Builder (building weekly lineup). See `plans/ffl-frontend-pages.md` for full page inventory.

## Tasks

### 1. Routing restructure
- [ ] FFL Home becomes `/` (app front door)
- [ ] AFL views move under `/afl/...`
- [ ] Navigation updated (FFL primary, AFL linked)

### 2. FFL Home page
- [ ] FFL ladder for current season
- [ ] Current round's matches with fantasy scores
- [ ] Round navigation
- [ ] Link to AFL section

### 3. FFL Round page
- [ ] All matches in round with scores
- [ ] Top fantasy scorers across the round
- [ ] Round navigation

### 4. FFL Match page (money shot)
- [ ] Head-to-head: two club rosters side by side
- [ ] Player details: name, FFL position, status, fantasy score
- [ ] Bench/sub/interchange indicators
- [ ] Club fantasy score totals

### 5. FFL Team Builder (money shot — stubbed)
- [ ] Layout with position slots and roster panel
- [ ] Display roster (30 players)
- [ ] Assign players to positions (local state only)
- [ ] Compare lineup arrangements
- [ ] No persistence yet — stub UI only

### 6. FFL Players page
- [ ] Player CRUD (create, edit, delete)
- [ ] Assign/remove players to/from club seasons

### 7. Backend wiring (end of sprint)
- [ ] Add `aflPlayerId` to FFL Player (domain + schema + migration)
- [ ] `setFFLLineup` mutation (batch upsert PlayerMatch)
- [ ] Roster query via GraphQL
- [ ] Wire Team Builder UI to real data

### 8. Playwright tests
- [ ] FFL Home tests
- [ ] FFL Round tests
- [ ] FFL Match tests
- [ ] FFL Team Builder tests
- [ ] FFL Players tests

## Out of scope
- Event subscription (AFL→FFL) — Phase 7
- Draft/trade mechanics — future phase
- Pulling AFL stats from external source — future phase
