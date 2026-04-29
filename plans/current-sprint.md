# Current Sprint — Phase 19: Graph Federation

**Sprint goal:** Adopt Apollo Federation so the frontend can traverse cross-service entity relationships (`FFLPlayerSeason → AFLPlayerSeason → AFLPlayer/AFLClub`, `FFLPlayerMatch → AFLPlayerMatch → stats`) in a single query. Replace the path-based Go gateway with Apollo Router.

ADR: ADR-013 (revised 2026-04-28)

---

## Step 1 — Apollo Router (gateway replacement)

- [x] Add Apollo Router to `dev/docker-compose.yml` on port `:4000`
- [x] Write `dev/router/supergraph.yaml` — compose from AFL (`:8080`) and FFL (`:8081`) subgraphs
- [x] CORS config in `dev/router/router.yaml`; `/health` built into Router
- [x] Go gateway updated: `/query` → Apollo Router; `/afl/query` and `/ffl/query` removed
- [x] `justfile`: `supergraph-compose` recipe added

## Step 2 — AFL subgraph

- [x] Add `AFLPlayerSeason @key(fields: "id")` type to AFL GraphQL schema
- [x] Add `aflPlayerSeason(id: ID!): AFLPlayerSeason` root query
- [x] Add `@key(fields: "id")` to `AFLPlayerMatch`
- [x] Implement entity resolvers: `FindAFLPlayerSeasonByID`, `FindAFLPlayerMatchByID`
- [x] Implement field resolvers: `AFLPlayerSeason.Player`, `.ClubSeason`, `.Matches`
- [x] sqlc: `FindPlayerSeasonsByIDs`, `FindPlayerMatchesByPlayerSeasonID`
- [x] `PlayerSeasonByID` dataloader added
- [x] AFL service mounts as federation subgraph (gqlgen federation plugin)

## Step 3 — FFL subgraph

- [x] Add stub `AFLPlayerSeason @key` and `AFLPlayerMatch @key` types to FFL schema
- [x] Add `aflPlayerSeason` field on `FFLPlayerSeason`
- [x] Add `aflPlayerMatch` and `aflPlayerMatchId` on `FFLPlayerMatch`
- [x] FFL entity reference resolvers return stub objects (router fetches from AFL)
- [x] `convertPlayerMatch` updated to populate `AflPlayerMatchID`
- [x] FFL service mounts as federation subgraph

## Step 4 — Frontend

- [x] Removed operation-name routing link
- [x] Apollo client points at single `/query` endpoint (via gateway → Router)
- [ ] Update `SquadView` to fetch player AFL club name via `FFLPlayerSeason.aflPlayerSeason.clubSeason.club.name`
- [ ] Verify existing queries still work end-to-end

## Step 5 — Tests + verification

- [x] AFL integration tests green
- [x] FFL integration tests green
- [ ] E2e test: SquadView shows AFL club name column
- [ ] All existing Playwright tests green
