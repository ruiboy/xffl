# XFFL — Fantasy Football League (the X makes it cool)

Multi-service fantasy football application bridging real AFL statistics with fantasy league scoring and search.

- **AFL** = Australian Football League (real match data)
- **FFL** = Fantasy Football League (fantasy teams, scoring, ladder)

**Tech stack:** Go, GraphQL, PostgreSQL, Zinc, Vue 3

## Architecture

![Logical View](doc/logical-view.png)

### Data Models

| AFL | FFL |
|-----|-----|
| ![AFL ERD](doc/erd-afl.png) | ![FFL ERD](doc/erd-ffl.png) |

PlantUML sources: [doc/logical-view.puml](doc/logical-view.puml), [doc/erd-afl.puml](doc/erd-afl.puml), [doc/erd-ffl.puml](doc/erd-ffl.puml)

See `ai/architecture/` for bounded contexts, service map, and principles.

## Getting Started

### Prerequisites

- Go 1.16+
- PostgreSQL 13+
- ZincSearch
- Node.js 16+

### Database

```bash
createdb xffl
psql -U postgres -d xffl -f infrastructure/postgres/migrations/001_create_afl_tables_up.sql
psql -U postgres -d xffl -f infrastructure/postgres/migrations/002_create_ffl_tables_up.sql

# Optional test data
psql -U postgres -d xffl -f infrastructure/postgres/test_data/insert_afl_data.sql
psql -U postgres -d xffl -f infrastructure/postgres/test_data/insert_ffl_data.sql
```

### Search Engine

```bash
brew tap zinclabs/tap && brew install zinclabs/tap/zincsearch
ZINC_FIRST_ADMIN_USER=admin ZINC_FIRST_ADMIN_PASSWORD=admin zincsearch
curl -u admin:admin -X PUT http://localhost:4080/api/index -d @infrastructure/zinc/xffl-index-config.json -H "Content-Type: application/json"
```

### Run Services

```bash
cd services/afl && go run cmd/server/main.go    # :8080
cd services/ffl && go run cmd/server/main.go    # :8081
cd services/search && go run cmd/server/main.go # :8082
cd gateway && go run main.go                    # :8090
cd frontend/web && npm install && npm run dev   # :3000
```

## Key Docs

| Doc | Purpose |
|-----|---------|
| [CLAUDE.md](CLAUDE.md) | Primary instructions for AI agents |
| [ai/architecture/](ai/architecture/) | Principles, service map, bounded contexts |
| [ai/plans/](ai/plans/) | Roadmap and current sprint |
| [ai/decisions/](ai/decisions/) | Architecture Decision Records |
