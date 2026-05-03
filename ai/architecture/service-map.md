# Service Map

| Service | Bounded Context | Port | API | Purpose |
|---------|----------------|------|-----|---------|
| AFL | AFL | 8080 | GraphQL | AFL clubs, players, match statistics |
| FFL | FFL | 8081 | GraphQL | Fantasy clubs, players, scoring, ladder |
| Search | — | 8082 | REST | Full-text search via Typesense |
| Gateway | — | 8090 | GraphQL | Single entry point for frontend; routes queries to AFL/FFL services, proxies search |

## Event Flow

```
AFL.PlayerMatchUpdated   → FFL (calculates fantasy score)
                         → Search (indexes player match)

AFL.PlayerSeasonUpdated  → Search (indexes player season)
                           [trigger not yet wired — Step 1 responsibility]

FFL.FantasyScoreCalculated → Search (indexes fantasy score)
```

## Infrastructure

- **Database:** PostgreSQL — separate schemas (`afl.*`, `ffl.*`)
- **Events:** PostgreSQL LISTEN/NOTIFY
- **Search engine:** Typesense (port 8108)
- **Frontend:** Vue 3 + Apollo Client (port 3000)

## Shared Packages (`shared/`)

- `database/` — connection utilities
- `events/` — dispatcher interface + PostgreSQL and in-memory implementations
