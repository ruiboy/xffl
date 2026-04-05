---
status: accepted
date: 2026-04-04
scope: architecture
enforceable: false
---

# ADR-013: Cross-Service Data Strategy — No Federation, CQRS Read/Write Split

## Context

The AFL and FFL GraphQL services are isolated. The frontend must issue separate queries to each and join client-side (e.g. SquadView fetches FFL squad data then AFL player stats separately). As the app grows — particularly player stats, averages, and historical match data — this client-side stitching will multiply.

The obvious architectural response is Apollo Federation: each service becomes a subgraph, the gateway composes a supergraph, and the frontend writes a single traversal query. We evaluated this path seriously.

In parallel, the long-term data model calls for a search index (Zinc → Elasticsearch) as the read layer for player stats and aggregations, with event-driven indexing from both AFL and FFL write operations (ADR-006, ADR-004).

## Decision

**Do not federate the graph. Use a CQRS read/write split instead.**

- **GraphQL services are write-model APIs** — they handle commands (enter stats, build squads, set teams) and serve structured domain data (seasons, rounds, matches, squads). They are not the authority for aggregated player stats at read time.
- **The search index is the read model for player stats** — averages, per-round history, rankings. The frontend queries the search index directly for this data, keyed by player/season IDs obtained from GraphQL.
- **Client-side stitching between GraphQL and search is intentional** — this is the expected pattern in CQRS, not a workaround. The two sides of the split are deliberately separate.

## Rationale

Federation solves a problem this architecture doesn't have in the long run. The core use case driving the federation discussion — "traverse FFL squad → AFL player season → AFL player match stats in one query" — dissolves once match stats live in the search index. Federating the graph would add infrastructure complexity (Apollo Router, `rover` CLI, subgraph entity resolvers) for a traversal pattern that will be superseded.

The CQRS split is also a better fit for the domain:
- Player stats are aggregated, denormalised, high-read — search index wins
- Squad composition, teams, match results are relational, transactional — GraphQL/PG wins
- The boundary between the two is clean and stable

## Consequences

- **Gateway routing stays path-based** — `/afl/query` and `/ffl/query` remain separate endpoints. Frontend routing link should be made explicit (operation-name map) rather than regex-based.
- **FFL service adds targeted resolvers** — e.g. `fflClubSeason(id: ID!)` so views don't over-fetch (load whole ladder for one club). These are standard GraphQL resolver improvements, not federation.
- **Search indexing expands** — AFL player match stats (per round, aggregated) are indexed into Zinc/Elasticsearch via events. SquadView and similar views query search for stats.
- **Federation remains an option** — if a genuine cross-graph traversal use case emerges that the search index cannot satisfy, federation can be revisited. The services are structured in a way that makes it feasible later.