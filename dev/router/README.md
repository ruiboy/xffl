# dev/router

Apollo Router configuration for local development.

## Files

| File | Status | Description |
|---|---|---|
| `router.yaml` | hand-written | Apollo Router config for local dev (subgraphs on `:8080`/`:8081`) |
| `router.test.yaml` | hand-written | Apollo Router config for e2e tests (subgraphs on `:8180`/`:8181`) |
| `supergraph.yaml` | hand-written | Rover composition config: subgraph names and SDL source URLs |
| `supergraph.graphql` | **generated** | Composed supergraph SDL — do not edit by hand |

## Regenerating the supergraph

Run whenever either subgraph schema changes:

```
just supergraph-compose
```

Requires the AFL (`:8080`) and FFL (`:8081`) services to be running, and the `rover` CLI installed (`~/.rover/bin/rover`).

The generated `supergraph.graphql` is committed so the Apollo Router Docker container can mount it without needing rover at runtime.
