# Developer Cookbook

Recipes and patterns for implementing changes across the stack. Read this before exploring the codebase.

## Service structure (identical for AFL and FFL)

```
services/{afl,ffl}/
  cmd/main.go                              → Wiring: pool → sqlcgen.Queries → repos → app.Queries/Commands → Resolver → server
  api/graphql/{query,mutation}.graphqls     → GraphQL schema (source of truth for API)
  gqlgen.yml                               → gqlgen config (models, resolver layout)
  sqlc.yaml                                → SQLC config (points at schema SQL + query SQL)
  internal/
    domain/                                → Structs, enums, repository interfaces. Zero dependencies.
    application/
      queries.go                           → Read-only operations. Holds all repo interfaces.
      commands.go                          → Write operations. Uses TxManager for transactions.
    infrastructure/postgres/
      sqlc/*.sql                           → Hand-written SQL queries (SQLC input)
      sqlcgen/                             → Generated Go from SQLC (DO NOT EDIT)
      repository.go                        → Implements domain repo interfaces, maps sqlcgen types ↔ domain types
      db.go                                → TxManager, transaction helpers
    interface/graphql/
      resolver.go                          → Resolver struct { Queries, Commands }
      query.resolvers.go                   → Query resolver implementations
      mutation.resolvers.go                → Mutation resolver implementations
      convert.go                           → Domain ↔ GraphQL type converters (toID, fromID, convertPlayer, etc.)
      generated.go                         → gqlgen generated (DO NOT EDIT)
      models_gen.go                        → gqlgen generated models (DO NOT EDIT)
```

## Recipe: Add a column to an existing table

Files to touch, in order:

1. **Schema** — `dev/postgres/init/{01_afl,02_ffl}_schema.sql` — add the column
2. **Domain** — `internal/domain/<entity>.go` — add field to struct (use `*int` for nullable)
3. **SQLC queries** — `internal/infrastructure/postgres/sqlc/<entity>.sql` — add column to SELECT lists and RETURNING clauses
4. **Generate SQLC** — run `cd services/{svc} && sqlc generate`
5. **Repository** — `internal/infrastructure/postgres/repository.go` — map the new sqlcgen field to domain field. Use helpers: `int32PtrToIntPtr`, `derefOr`, `positionPtr`, etc.
6. **GraphQL schema** — `api/graphql/query.graphqls` — add field to the relevant type
7. **Generate gqlgen** — run `cd services/{svc} && go run github.com/99designs/gqlgen generate`
8. **Resolver** — `internal/interface/graphql/query.resolvers.go` — populate the new field (may need converter update in `convert.go`)
9. **Seed data** — `dev/postgres/seed/*.sql` — add values for the new column
10. **Reset DB** — `just dev-reset && just dev-up && just dev-seed`

## Recipe: Add a new GraphQL query

1. **Schema** — add query to `api/graphql/query.graphqls` (and any new types)
2. **Generate gqlgen** — creates a panic stub in `query.resolvers.go`
3. If it needs new data access:
   - Add SQLC query → `sqlc generate` → add repo method → add domain interface method → add application query method
4. **Implement resolver** — replace the panic stub, use `fromID`/`toID` for ID conversion
5. **Frontend** — add gql query in `frontend/web/src/features/{ffl,afl}/api/queries.ts`

## Recipe: Add a new frontend page

1. **View component** — `frontend/web/src/features/{ffl,afl}/views/<Name>View.vue`
2. **Route** — `frontend/web/src/app/router.ts` — add route with `props: true`
3. **Query** — `features/{ffl,afl}/api/queries.ts` — add GraphQL query
4. **Pattern**: `useQuery` from `@vue/apollo-composable`, loading/error/data template guards, Tailwind styling

## Code generation commands

```bash
cd services/afl && sqlc generate                           # Regenerate SQLC (after changing .sql queries or schema)
cd services/afl && go run github.com/99designs/gqlgen generate  # Regenerate gqlgen (after changing .graphqls)
cd services/ffl && sqlc generate
cd services/ffl && go run github.com/99designs/gqlgen generate
```

Always run `sqlc generate` before `gqlgen generate` if both schema SQL and GraphQL changed.

## Cross-service queries (frontend)

The Apollo client (`frontend/web/src/app/apollo.ts`) routes by **operation name** using an explicit set `FFL_OPERATIONS`. Operations in the set go to FFL (`/ffl/query`); everything else goes to AFL (`/afl/query`).

**When adding a new FFL operation**, add its name to `FFL_OPERATIONS` in `apollo.ts`. Do not rely on naming conventions — the set is the source of truth (see ADR-008).

**A single GraphQL operation cannot span both services.** For cross-service data (e.g. FFL squad + AFL stats), issue two separate queries and join in the component. Pattern:

```ts
const { result: fflResult } = useQuery(FFL_QUERY, ...)
const ids = computed(() => /* extract AFL IDs from fflResult */)
const { result: aflResult } = useQuery(AFL_QUERY, () => ({ ids: ids.value }), () => ({ enabled: ids.value.length > 0 }))
// join in a computed
```

## Recipe: Add a paginated list field (Connection pattern)

See ADR-014. Use this for any list that may grow beyond ~50 items.

1. **Schema** — define the connection and filter types in `query.graphqls`:
```graphql
type <Resource>Connection {
  nodes: [<Resource>!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

input <Resource>Filter {
  # add filter fields as needed
}
```
`PageInfo` is already defined in `api/graphql/common.graphqls` — do not redefine it.

2. **Field declaration** — add to the parent type:
```graphql
items(first: Int, after: String, filter: <Resource>Filter): <Resource>Connection!
```

3. **gqlgen.yml** — mark the field as a resolver:
```yaml
ParentType:
  fields:
    items: { resolver: true }
```

4. **Generate** — `go run github.com/99designs/gqlgen generate`

5. **Implement resolver** — until real pagination is needed, return all items with `hasNextPage: false`:
```go
return &ResourceConnection{
  Nodes:      nodes,
  PageInfo:   &PageInfo{HasNextPage: false},
  TotalCount: len(nodes),
}, nil
```

## Repository type mapping helpers

Both services use these in `repository.go`:
- `int32PtrToIntPtr` / `intPtrToInt32Ptr` — nullable int conversion between sqlcgen (int32) and domain (int)
- `derefOr(p *int32) int` — dereference pointer or return 0
- `toID(int) string` / `fromID(string) (int, error)` — in convert.go, for GraphQL ID ↔ int

## Frontend conventions

- **Styling**: Tailwind with semantic tokens — `text-text-muted`, `bg-surface-raised`, `border-border`, `bg-active`, `text-active-text`
- **Tables**: `overflow-x-auto` wrapper, `w-full text-sm`, `border-b border-border` headers, `tabular-nums` for numbers
- **Components**: Vue 3 `<script setup>`, TypeScript, feature-folder structure (`features/{afl,ffl}/{api,components,views}`)
- **Numbers**: `.toFixed(1)` for averages, `tabular-nums` class for alignment

## Recipe: Add an integration (external data source)

See `ai/architecture/integrations.md` for the production adapter pattern (ACL, outbound ports, secondary adapters, cache policy) and `ai/architecture/historical-import.md` for the historical import pattern (two-phase reconcile → import, xref tables, fuzzy player matching).

**Production adapter** (recurring, scheduled):
1. Define outbound port interface in `internal/application/`
2. Create adapter package `internal/infrastructure/<source>/`
3. Add `xref_<source>_<entity>` table to `dev/postgres/init/<n>_<service>_integrations.sql`
4. Wire adapter → use case → DB writes → domain events in `cmd/ingest/main.go`

**Historical import** (one-time dev tool):
1. Add `xref_<source>_<entity>` table to `dev/postgres/init/<n>_<service>_integrations.sql`
2. Build `dev/import/<source>/main.go` with `--reconcile` and import modes
3. Run reconciliation, review and commit `reconcile.csv`, then import

## Testing

See `ai/architecture/testing.md` for Go testing conventions (stack, patterns, naming).

- **Frontend type check**: `cd frontend/web && npx vue-tsc --noEmit`
- **E2E**: `just test-e2e` (self-contained — starts its own isolated stack)
