---
status: accepted
date: 2026-04-28
revised: 2026-04-28
scope: architecture
enforceable: true
rules:
  - "AFL and FFL GraphQL services are Apollo Federation subgraphs — no standalone /afl/query or /ffl/query exposure to the frontend"
  - "cross-graph entity traversal uses federation entity resolvers — not ad-hoc Twirp RPCs"
  - "Apollo Router is the single GraphQL entry point for the frontend"
  - "aggregated/search reads (averages, rankings, season history) use Typesense directly, keyed by IDs from the federated graph"
  - "Twirp (ADR-018) remains valid for non-graph blocking RPCs (e.g. import player resolution)"
---

# ADR-013: Cross-Service Data Strategy — Federation for Structural Traversal, CQRS for Aggregated Reads

## History

**2026-04-04 (original):** No federation; CQRS read/write split. Rationale: all stats traversal would dissolve into the Typesense search index, making federation unnecessary.

**2026-04-28 (this revision):** Adopt Apollo Federation. The original decision drew the Typesense boundary too broadly. Three concrete traversal patterns have emerged that are structural entity lookups, not search queries. The Twirp approach would require 3+ new RPCs and batchers for patterns GraphQL federation handles naturally.

---

## Context

The original no-federation decision rested on the premise that cross-graph traversal needs ("traverse FFL squad → AFL player → AFL match stats") would dissolve once the Typesense read model was built. That premise was partially correct but missed a category of data.

Three cross-graph traversal patterns have since been identified:

| Pattern | Example | Type |
|---|---|---|
| Identity | `FFLPlayerSeason → AFLPlayerSeason → player name + club name` | Structural |
| Per-match stats | `FFLPlayerMatch → AFLPlayerMatch → kicks/handballs` | Structural |
| Season history | `FFLPlayerSeason → [AFLPlayerMatch] → kicks per round` | Aggregated |

Patterns 1 and 2 are **structural entity traversal** — a relational lookup on a specific entity. Typesense is not the right tool; these are not search or aggregation queries. Pattern 3 is genuinely aggregated and remains Typesense territory.

The DB FK links for patterns 1 and 2 already exist (`ffl.player_season.afl_player_season_id`, `ffl.player_match.afl_player_match_id`). The gap is purely at the graph layer.

With two structural traversal patterns and more plausible future ones, the Twirp approach (ADR-018) would require multiple new RPCs and dataloadgen batchers per pattern, each re-shaping AFL data under FFL's schema. At that scale, Twirp becomes a manual, less ergonomic reimplementation of federation.

## Decision

**Adopt Apollo Federation. Preserve the CQRS split for aggregated reads.**

- **AFL and FFL services become Apollo Federation subgraphs.** Each exposes a subgraph-compatible GraphQL schema with `@key` directives on entities that other services reference.
- **Apollo Router replaces the path-based gateway** (superseding ADR-008's "stay on path-based routing"). The frontend hits one endpoint and writes natural traversal queries. Supersedes ADR-008 gateway routing.
- **Federation is for structural entity traversal only.** It is not the read model for aggregated stats.
- **Typesense remains the read model for aggregated/search reads** — season averages, rankings, top scorers. The frontend stitch between the federated graph (for IDs and structure) and Typesense (for aggregated stats) is intentional.
- **Twirp (ADR-018) remains valid** for genuinely blocking non-graph RPCs — e.g. import flows that need to resolve an AFL player ID before writing FFL records.

## Subgraph Architecture

### AFL subgraph

Entities to expose with `@key`:

```graphql
type AFLPlayerSeason @key(fields: "id") {
  id: ID!
  player: AFLPlayer!
  clubSeason: AFLClubSeason!
  matches: [AFLPlayerMatch!]!
}

type AFLPlayerMatch @key(fields: "id") {
  id: ID!
  playerSeasonId: ID!
  player: AFLPlayer!
  status: String!
  kicks: Int!
  handballs: Int!
  marks: Int!
  hitouts: Int!
  tackles: Int!
  goals: Int!
  behinds: Int!
  disposals: Int!
  score: Int!
}
```

Entity resolvers batch by IDs using the existing dataloadgen pattern (ADR-017).

### FFL subgraph

Reference AFL entities by ID; the router resolves them via the AFL subgraph entity resolver:

```graphql
type FFLPlayerSeason {
  id: ID!
  player: FFLPlayer!
  clubSeasonId: ID!
  aflPlayerSeasonId: ID
  aflPlayerSeason: AFLPlayerSeason   # resolved via federation
}

type FFLPlayerMatch {
  id: ID!
  playerSeasonId: ID!
  player: FFLPlayer!
  score: Int!
  aflPlayerMatchId: ID
  aflPlayerMatch: AFLPlayerMatch      # resolved via federation
}
```

FFL adds stub reference resolvers that return the entity key fields only; the router handles the cross-subgraph fetch.

### Gateway

Apollo Router replaces `services/gateway`. It runs as a Docker container in `dev/docker-compose.yml`, configured with a supergraph schema composed from both subgraphs at startup.

## Rationale

- **The threshold is crossed.** Two confirmed structural traversal patterns plus likely future ones (e.g. `FFLMatch → AFLMatch → venue/result`) mean the Twirp approach would compound indefinitely.
- **AFL graph already has `AFLPlayerMatch` with full stat fields.** Adding `AFLPlayerSeason` and `@key` directives is incremental, not a redesign.
- **DB FK links exist.** `afl_player_season_id` and `afl_player_match_id` are already in `ffl.*` tables. This is purely a graph-layer change.
- **Frontend simplifies.** One endpoint, one Apollo client link — no operation-name routing map to maintain.
- **Twirp and events are unaffected.** Federation handles graph traversal; Twirp handles import-time blocking lookups; events handle async cross-service notification.

## Consequences

- **Apollo Router** — added to Docker Compose as the single GraphQL entry point (`:8090`). Replaces the Go gateway service for GraphQL traffic. CORS and `/health` move to the Router config.
- **AFL service** — add `AFLPlayerSeason` type and entity resolver; add `@key` to `AFLPlayerMatch`; mount Apollo Federation subgraph handler.
- **FFL service** — add `aflPlayerSeason` and `aflPlayerMatch` fields with reference resolvers; expose as federation subgraph.
- **Frontend** — single Apollo client endpoint; remove operation-name routing link.
- **ADR-008 superseded** — path-based routing decision no longer applies to GraphQL. Health check endpoint pattern is retained.
- **ADR-015 (Typesense) unchanged** — aggregated read model is unaffected.
- **ADR-017 (dataloadgen) unchanged** — entity resolvers use the same batch-by-IDs pattern.
- **ADR-018 (Twirp) unchanged** — import-time blocking RPCs are a distinct pattern; federation does not replace them.