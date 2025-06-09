# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Backend (Go)
- Start backend server: `cd backend && go run cmd/server/main.go`
- Generate GraphQL code: `cd backend && go run github.com/99designs/gqlgen generate`
- Build: `cd backend && go build -o bin/server cmd/server/main.go`

### Frontend (Vue.js)
- Install dependencies: `cd frontend && npm install`
- Start dev server: `cd frontend && npm run dev`
- Build for production: `cd frontend && npm run build && npm run preview`

### Database
- Run migrations: `psql -U postgres -d gffl -f backend/internal/adapters/persistence/migrations/001_create_ffl_tables_up.sql`
- Insert test data: `psql -U postgres -d gffl -f backend/internal/adapters/persistence/test_scripts/insert_test_clubs.sql`

## Architecture

This is a full-stack fantasy football league application with Clean Architecture + Hexagonal Architecture principles:

- **Backend**: Go with GraphQL API using gqlgen, GORM for database ORM, PostgreSQL database
- **Frontend**: Vue 3 + TypeScript + Vite, Apollo Client for GraphQL, PrimeVue UI components
- **Database Schema**: Uses `ffl` schema with `club` and `player` tables, managed via SQL migrations

### Backend Architecture

#### Clean Architecture Layers:
- **Domain Layer** (`internal/domain/`): Pure business logic, entities, value objects, domain events
- **Application Layer** (`internal/application/`): Use cases and application services that orchestrate domain operations
- **Interface Adapters** (`internal/adapters/`): Adapters for external systems (GraphQL, REST, Database, PubSub)
- **Infrastructure** (`internal/infrastructure/`): Framework and tools configuration

#### Hexagonal Architecture Ports:
- **Input Ports** (`internal/ports/in/`): Interfaces defining how external systems can interact with the application
- **Output Ports** (`internal/ports/out/`): Interfaces defining how the application interacts with external systems

### Key Backend Components

- `api/graphql/schema.graphqls`: GraphQL schema definition (easily discoverable at top level)
- `cmd/server/main.go`: Application entry point and server setup
- `internal/adapters/graphql/`: GraphQL resolvers (input adapters)
- `internal/adapters/persistence/`: Database models and repositories (output adapters)
- `internal/domain/`: Business entities (Club, Player, etc.) and domain logic
- `internal/application/`: Use cases for business operations
- `internal/infrastructure/`: Database connections, HTTP server configuration
- `internal/ports/`: Interface definitions for input and output ports
- Database uses environment variables (DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT)

### Key Frontend Components

- Uses Apollo Client for GraphQL queries/mutations
- Vue Router for navigation between Home and Players views
- PrimeVue for UI components
- State management with Pinia

### Development Workflow

1. After modifying GraphQL schema in `api/graphql/schema.graphqls`, always run `cd backend && go run github.com/99designs/gqlgen generate`
2. Backend runs on :8080 with GraphQL playground at root and API at /query
3. Frontend runs on :3000 with CORS configured for localhost communication
4. Database changes require manual SQL migration execution in `internal/adapters/persistence/migrations/`
5. Business logic goes in `internal/domain/` (entities) and `internal/application/` (use cases)
6. External integrations go in `internal/adapters/` with corresponding interfaces in `internal/ports/`
