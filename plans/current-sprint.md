# Current Sprint — Phase 19: Graph Federation

**Sprint goal:** Adopt Apollo Federation so the frontend can traverse cross-service entity relationships (`FFLPlayerSeason → AFLPlayerSeason → AFLPlayer/AFLClub`, `FFLPlayerMatch → AFLPlayerMatch → stats`) in a single query. Replace the path-based Go gateway with Apollo Router.

ADR: ADR-013 (revised 2026-04-28)

---

## Step 1 — Apollo Router (gateway replacement)

- [ ] Add Apollo Router to `dev/docker-compose.yml` on port `:8090` (replacing the Go gateway container)
- [ ] Write `dev/router/supergraph.yaml` — compose from AFL (`:8080`) and FFL (`:8081`) subgraphs
- [ ] Migrate CORS config and `/health` endpoint to Router config
- [ ] Remove `services/gateway` Go service (or keep as stub — decide when Router is confirmed working)
- [ ] Update `justfile` recipes that reference the gateway

## Step 2 — AFL subgraph

- [ ] Add `AFLPlayerSeason` type to AFL GraphQL schema:
  - `id: ID!`, `player: AFLPlayer!`, `clubSeason: AFLClubSeason!`, `matches: [AFLPlayerMatch!]!`
- [ ] Add `aflPlayerSeason(id: ID!): AFLPlayerSeason` root query
- [ ] Add `@key` federation directive to `AFLPlayerSeason` and `AFLPlayerMatch`
- [ ] Implement entity resolver: batch `AFLPlayerSeason` by IDs (dataloadgen, ADR-017)
- [ ] Implement entity resolver: batch `AFLPlayerMatch` by IDs (dataloadgen)
- [ ] Add DB query + sqlc for `FindPlayerSeasonsByIDs`
- [ ] Mount AFL handler as a federation subgraph (gqlgen federation plugin)
- [ ] Integration tests for new resolvers

## Step 3 — FFL subgraph

- [ ] Add `aflPlayerSeason: AFLPlayerSeason` field to `FFLPlayerSeason` schema type
- [ ] Add `aflPlayerMatch: AFLPlayerMatch` field to `FFLPlayerMatch` schema type
- [ ] Expose `afl_player_match_id` as `aflPlayerMatchId: ID` on `FFLPlayerMatch` (it exists in DB, not yet in schema)
- [ ] Implement reference resolvers for both fields (return entity key only — Router fetches from AFL subgraph)
- [ ] Mount FFL handler as a federation subgraph
- [ ] Integration tests

## Step 4 — Frontend

- [ ] Remove the Apollo operation-name routing link (`apolloLink.ts` or equivalent)
- [ ] Point Apollo client at single Router endpoint (`:8090/query`)
- [ ] Update `SquadView` to fetch player AFL club name via `FFLPlayerSeason.aflPlayerSeason.clubSeason.club.name`
- [ ] Verify existing queries still work end-to-end

## Step 5 — Tests + verification

- [ ] E2e test: SquadView shows AFL club name column
- [ ] All existing Playwright tests green
- [ ] All Go integration tests green
