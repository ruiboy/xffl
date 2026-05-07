# Route Migration Plan

Migrate all leaf page routes from hierarchical (`/seasons/:sid/rounds/:rid`) to flat entity IDs (`/rounds/:id`).
Edit/mutation screens use the `/edit` suffix; drop the `/admin/` prefix (access control belongs in route guards, not URLs).

---

## Affected routes

Route names are updated alongside paths — rename every reference in the same pass.

| Old name | New name | Old path | New path | Complexity |
|---|---|---|---|---|
| `ffl-round` | (unchanged) | `/ffl/seasons/:sid/rounds/:rid` | `/ffl/rounds/:id` | Low |
| `ffl-match` | (unchanged) | `/ffl/seasons/:sid/matches/:mid` | `/ffl/matches/:id` | Low |
| `afl-round` | (unchanged) | `/afl/seasons/:sid/rounds/:rid` | `/afl/rounds/:id` | Low |
| `afl-match` | (unchanged) | `/afl/seasons/:sid/matches/:mid` | `/afl/matches/:id` | Low |
| `afl-admin-match` | `afl-match-edit` | `/admin/afl/seasons/:sid/matches/:mid` | `/afl/matches/:id/edit` | Low |
| `ffl-squad` | `ffl-club-season` | `/ffl/seasons/:sid/clubs/:cid/squad` | `/ffl/club-seasons/:id` | Medium |
| `ffl-team-builder` | `ffl-club-match-edit` | `/ffl/seasons/:sid/rounds/:rid/team-builder` | `/ffl/club-matches/:id/edit` | High |

---

## Phase 1 — Simple routes (drop seasonId from path)

These routes only need `seasonId` removed from the URL; the view components use it only as a passthrough to GraphQL queries that can derive it from the leaf entity.

### 1. `router.ts`

For each route: rename path + add redirect from old path.

```ts
// New
{ path: '/ffl/rounds/:roundId', name: 'ffl-round', ... }
{ path: '/ffl/matches/:matchId', name: 'ffl-match', ... }
{ path: '/afl/rounds/:roundId', name: 'afl-round', ... }
{ path: '/afl/matches/:matchId', name: 'afl-match', ... }
{ path: '/afl/matches/:matchId/edit', name: 'afl-match-edit', ... }

// Redirects (keep until no bookmarks exist)
{ path: '/ffl/seasons/:seasonId/rounds/:roundId', redirect: to => ({ name: 'ffl-round', params: { roundId: to.params.roundId } }) }
{ path: '/ffl/seasons/:seasonId/matches/:matchId', redirect: to => ({ name: 'ffl-match', params: { matchId: to.params.matchId } }) }
// ... same for afl-round, afl-match, afl-admin-match
```

### 2. View component prop changes

Each view currently accepts `seasonId` as a prop but only uses it to pass to queries that could derive season from the entity. Remove `seasonId` from `defineProps`; fetch season context via the leaf entity query.

- `services/ffl/views/RoundView.vue` — remove `seasonId` prop; season ID is in `round.season.id`
- `services/ffl/views/MatchView.vue` — remove `seasonId` prop; season/round IDs are in `match.round.season.id` etc.
- `services/afl/views/RoundView.vue` — remove `seasonId` prop
- `services/afl/views/MatchView.vue` — remove `seasonId` prop
- `services/afl/views/AdminMatchView.vue` — remove `seasonId` prop

Check each view's queries: if they pass `seasonId` as a variable, update them to read it from the fetched entity instead.

### 3. Update all links

Files that reference the above route names (from grep):

**`ffl-round`**
- `features/ffl/components/RoundNav.vue:6` — drop `seasonId` from params
- `features/ffl/views/RoundView.vue:109` — drop `seasonId`
- `features/ffl/views/MatchView.vue:84` — drop `seasonId`
- `features/ffl/views/TeamBuilderView.vue:448` — drop `seasonId`
- `features/data-ops/views/DataOpsView.vue:199` — already uses `liveSeasonId`, just drop it

**`ffl-match`**
- `features/ffl/views/RoundView.vue:26` — drop `seasonId`
- `features/ffl/views/TeamBuilderView.vue:453` — drop `seasonId`

**`afl-round`**
- `features/afl/components/RoundNav.vue:6` — drop `seasonId`
- `features/ffl/views/RoundView.vue:109` — drop `seasonId`
- `features/ffl/views/MatchView.vue:97` — drop `seasonId`
- `features/afl/views/AdminMatchView.vue:55` — drop `seasonId`
- `features/data-ops/views/DataOpsView.vue:35` — drop `aflSeasonId`

**`afl-match`**
- `features/afl/views/RoundView.vue:25` — drop `seasonId`
- `features/afl/views/MatchView.vue:77` — drop `seasonId`
- `features/ffl/utils/aflPlayerMatch.ts:42` — drop `seasonId`
- `features/data-ops/views/DataOpsView.vue:67` — drop `aflSeasonId`

**`afl-match-edit`** (was `afl-admin-match`)
- No external links found (only defined in router + used internally in AdminMatchView)

---

## Phase 2 — `ffl-squad` → `ffl-club-season` at `/ffl/club-seasons/:id`

**What changes:** `SquadView` currently takes `seasonId` + `clubId`. The new route passes `clubSeasonId` (the `ffl.club_season` PK) directly.

### Schema addition
Add `fflClubSeason(id: ID!): FFLClubSeason` to the FFL query root + resolver. The existing compound-key field `fflClubSeason(seasonId: ID!, clubId: ID!)` can be removed once no callers remain (audit at end of phase).

### Requirements
- All links supplying `{ name: 'ffl-squad', params: { seasonId, clubId } }` need to supply `{ name: 'ffl-club-season', params: { clubSeasonId } }` instead.
- The callers need a `clubSeasonId` available at link-generation time. Audit each:
  - `App.vue:39` — has `selectedClubId` (club ID, not club season ID); needs `selectedClubSeasonId` from `useFflState` or a query
  - `features/ffl/components/LadderTable.vue:26` — has `entry.club.id`; needs `entry.id` (the club_season row)
  - `features/ffl/views/MatchView.vue:28` — has `side.clubMatch.club.id`; needs `side.clubMatch.clubSeasonId`
  - `features/ffl/views/TeamBuilderView.vue:33` — has `selectedClubSeason.club.id`; can use `selectedClubSeason.id` directly
- `SquadView.defineProps` changes to `{ clubSeasonId: string }`.
- Queries in SquadView that currently use `seasonId` + `clubId` switch to querying by `clubSeasonId`.

---

## Phase 3 — `ffl-team-builder` → `ffl-club-match-edit` at `/ffl/club-matches/:id/edit`

TeamBuilderView currently takes `seasonId` + `roundId` as props and reads `selectedClubId` from shared state. The new route passes a `clubMatchId` (the `ffl.club_match` PK).

### Schema addition
Add `fflClubMatch(id: ID!): FFLClubMatch` to the FFL query root + resolver.

### What needs to change in TeamBuilderView
1. `defineProps` changes to `{ clubMatchId: string }`.
2. On load: query `fflClubMatch(id)` to derive `roundId`, `seasonId`, and `clubId` — then call `setClub(clubId)` so shared state is correct.
3. Round-based queries (`GET_FFL_ROUND`) switch to using the derived `roundId`.
4. Prev/next round navigation currently links to `{ name: 'ffl-team-builder', params: { seasonId, roundId: prevRound.id } }`. With the new route, prev/next need the club's `clubMatch.id` for the adjacent round — look it up from the round query result (which already returns all club matches).
5. Breadcrumbs use derived `roundId`.

### What needs to change in callers
- **`App.vue:30`** — remove the team builder nav link entirely. The team builder is a contextual action; in-page links from the round/match/data-ops pages will always have `clubMatchId` naturally from their loaded data. No `liveClubMatchId` needed in shared state.
- `features/ffl/views/SquadView.vue:11` — links to team builder with `seasonId + liveRoundId`; after migration replace with the club's `clubMatchId` for the live round, available from the squad query result.
- `features/ffl/views/MatchView.vue:35` — has `round.id`; needs the user's club_match ID for that round, available from the match data already loaded.
- `features/data-ops/views/DataOpsView.vue` — `row.clubMatchId` already available; link directly to `/ffl/club-matches/:id/edit` per row (replaces the `setClub` workaround).

---

## Suggested order

1. **Phase 1 first** — purely mechanical, verifiable in isolation, unblocks cleaning up DataOpsView links immediately.
2. **Phase 2 (squad)** — medium; mostly link + query changes, minimal new data requirements.
3. **Phase 3 (team builder)** — do last; requires adding `liveClubMatchId` to shared state and reworking TeamBuilderView queries.