---
status: accepted
date: 2026-04-04
scope: interface
enforceable: true
rules:
  - "list fields that may grow use the Connection pattern"
  - "pagination is cursor-based only — no offset/limit"
  - "PageInfo is defined once in api/graphql/common.graphqls per service"
  - "Connection types are named <Resource>Connection with nodes, pageInfo, totalCount"
---

# ADR-014: GraphQL Pagination Standard — Cursor-Based Connection Pattern

## Context

As FFL list fields were introduced (e.g. `FFLClubSeason.players`), the question arose of how to handle pagination. Options considered: offset/limit, cursor-based, or no pagination (return all).

## Decision

Use cursor-based pagination with the Connection pattern for any list field that may grow.

**Query pattern:**
```graphql
query ($first: Int!, $after: String) {
  <resource>(first: $first, after: $after) {
    nodes { ... }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

**Type shape:**
```graphql
type PageInfo {
  hasNextPage: Boolean!
  endCursor: String
}

type <Resource>Connection {
  nodes: [<Resource>!]!
  pageInfo: PageInfo!
  totalCount: Int!
}
```

**Rules:**
- `PageInfo` is defined once in `api/graphql/common.graphqls` (per service) — never duplicated
- Connection fields accept `first: Int, after: String` arguments even before real pagination is implemented
- Cursors are opaque to clients (base64-encoded stable sort key, e.g. id)
- Always return `nodes`, `pageInfo`, `totalCount`
- Forward pagination only (`first` + `after`); no `last`/`before`
- No `edges` unless explicitly required

## Rationale

- **Why cursor-based:** Stable under concurrent inserts/deletes; offset pagination skips or duplicates rows when the underlying data changes
- **Why adopt the shape early:** Avoids breaking schema changes later; the shape is additive — clients can ignore `pageInfo` until they need it
- **Why not return all:** Defensive default for fields that may grow (squads, match history, stat lists)

## When to apply

Apply to any list field where the count is unbounded or expected to grow beyond ~50 items. Small fixed lists (e.g. `positions`, `rounds` in a season) may stay as plain arrays.

## Implementation notes

- Until real pagination is needed, resolvers return all items with `hasNextPage: false` and `endCursor: nil`
- When implementing real cursors: fetch `first + 1` items; set `hasNextPage = len > first`; encode last returned item's stable key as `endCursor`
- `PageInfo` lives in `common.graphqls` in each service — AFL and FFL each define their own (no cross-service schema sharing per ADR-002)
