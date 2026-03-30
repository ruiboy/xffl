# Service Map

| Service | Bounded Context | Port | API | Purpose |
|---------|----------------|------|-----|---------|
| AFL | AFL | 8080 | GraphQL | AFL clubs, players, match statistics |
| FFL | FFL | 8081 | GraphQL | Fantasy clubs, players, scoring, ladder |
| Search | — | 8082 | REST | Full-text search via Zinc |
| Gateway | — | 8090 | GraphQL | Single entry point for frontend; routes queries to AFL/FFL services, proxies search |

## Event Flow

```
AFL.PlayerMatchUpdated → FFL (calculates fantasy score)
                       → Search (indexes player match)

FFL.FantasyScoreCalculated → Search (indexes/updates player)
```

## Infrastructure

- **Database:** PostgreSQL — separate schemas (`afl.*`, `ffl.*`)
- **Events:** PostgreSQL LISTEN/NOTIFY
- **Search engine:** Zinc (port 4080)
- **Frontend:** Vue 3 + Apollo Client (port 3000)

## Shared Packages (`shared/`)

- `database/` — connection utilities
- `events/` — dispatcher interface + PostgreSQL and in-memory implementations
