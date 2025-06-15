# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Services (Go)
- Start FFL service: `cd services/ffl && go run cmd/server/main.go` (port 8080)
- Start AFL service: `cd services/afl && PORT=8081 go run cmd/server/main.go` (port 8081)
- Generate GraphQL code: `cd services/ffl && go run github.com/99designs/gqlgen generate`
- Build service: `cd services/ffl && go build -o bin/server cmd/server/main.go`

### Environment Variables
- `EVENT_DB_URL` - PostgreSQL connection string for cross-service events (default: "user=postgres dbname=xffl sslmode=disable")
- `PORT` - Service port (defaults: FFL=8080, AFL=8081, Gateway=8090)

### Gateway (Go)
- Start gateway: `cd gateway && go run main.go` (port 8090)
- No code generation needed - pure Go standard library

### Frontend (Vue.js)
- Install dependencies: `cd frontend && npm install`
- Start dev server: `cd frontend && npm run dev` (port 3000)
- Build for production: `cd frontend && npm run build && npm run preview`

### Database
- Run AFL migrations: `psql -U postgres -d xffl -f infrastructure/postgres/migrations/001_create_afl_tables_up.sql`
- Run FFL migrations: `psql -U postgres -d xffl -f infrastructure/postgres/migrations/002_create_ffl_tables_up.sql`
- Insert AFL test data: `psql -U postgres -d xffl -f infrastructure/postgres/test_data/insert_afl_data.sql`
- Insert FFL test data: `psql -U postgres -d xffl -f infrastructure/postgres/test_data/insert_ffl_data.sql`

## Architecture

This is a multi-service fantasy football league application with Clean Architecture + Hexagonal Architecture principles:

- **Services**: Independent Go microservices (AFL, FFL) with GraphQL APIs using gqlgen
- **Gateway**: Simple Go proxy service that routes GraphQL requests to appropriate backend services
- **Frontend**: Vue 3 + TypeScript + Vite, Apollo Client for GraphQL, PrimeVue UI components
- **Database**: PostgreSQL with separate schemas (`afl`, `ffl`), managed via SQL migrations

### Service Architecture

Each service follows Clean Architecture + Hexagonal Architecture:

#### Clean Architecture Layers:
- **Domain Layer** (`services/*/internal/domain/`): Pure business logic, entities, value objects
- **Application Layer** (`services/*/internal/application/`): Use cases and application services
- **Interface Adapters** (`services/*/internal/adapters/`): GraphQL resolvers, persistence adapters
- **Infrastructure** (`pkg/`): Shared database connections, configuration

#### Dependency Inversion Pattern:
We follow strict dependency inversion to keep the domain pure:
- **Domain entities**: Pure structs (e.g., `afl.Club`, `afl.PlayerMatch`) with JSON tags only
- **Database entities**: Separate structs (e.g., `ClubEntity`, `PlayerMatchEntity`) with GORM tags
- **Entity mapping**: Repository methods convert database ↔ domain entities
- **Interface definition**: Repository interfaces defined in `ports/out`, implemented in `adapters/persistence`

This ensures domain entities have zero infrastructure dependencies and can be tested in isolation.

#### Hexagonal Architecture Ports:
- **Input Ports** (`services/*/internal/ports/in/`): Service interfaces (use cases)
- **Output Ports** (`services/*/internal/ports/out/`): Repository and external service interfaces

### Gateway Architecture

The gateway provides a unified GraphQL endpoint using simple string-based routing:

- **Location**: `gateway/main.go` (single file, ~150 lines)
- **Dependencies**: Go standard library only
- **Routing Logic**: Routes based on presence of "afl" or "ffl" in query text
- **CORS**: Configured for frontend at localhost:3000
- **Health Check**: Available at `/health` endpoint

#### Gateway Routing:
- Queries containing `afl` → AFL service (localhost:8081)
- Queries containing `ffl` → FFL service (localhost:8080)
- Queries containing `_gateway` → Gateway metadata (handled locally)
- All other queries → FFL service (default)

### Key Components

#### Services:
- `services/*/api/graphql/schema.graphqls`: GraphQL schema definition
- `services/*/cmd/server/main.go`: Service entry point and server setup
- `services/*/internal/adapters/graphql/`: GraphQL resolvers (input adapters)
- `services/*/internal/adapters/persistence/`: Database entities and repositories
  - Contains database entities with GORM annotations (e.g., `ClubEntity`, `PlayerMatchEntity`)
  - Implements entity mapping between database and domain models
- `services/*/internal/domain/`: Pure business entities and domain logic
  - Contains pure domain entities with no infrastructure dependencies
- `services/*/internal/application/`: Use cases and business operations
- `services/*/internal/ports/`: Interface definitions for clean architecture

#### Gateway:
- `gateway/main.go`: Complete gateway implementation
- `gateway/go.mod`: Minimal dependencies (Go standard library only)

#### Frontend:
- Uses Apollo Client pointing to gateway (localhost:8090)
- Vue Router for navigation between views
- PrimeVue for UI components
- State management with Pinia

### Development Workflow

1. **Start Services**: Run AFL and FFL services on ports 8081 and 8080
2. **Start Gateway**: Run gateway on port 8090 to proxy requests
3. **Start Frontend**: Run Vue dev server on port 3000, configured to use gateway
4. **GraphQL Changes**: Modify schema in `services/*/api/graphql/schema.graphqls`, then run gqlgen generate
5. **Database Changes**: Create SQL migrations in `services/*/internal/adapters/persistence/migrations/`
6. **Business Logic**: Add domain entities in `internal/domain/` and use cases in `internal/application/`

### Request Flow

Frontend (3000) → Gateway (8090) → AFL Service (8081) or FFL Service (8080) → PostgreSQL

The gateway routes requests based on simple string matching in the GraphQL query text, providing a unified API surface while keeping services independent.

### Environment Variables

Services use environment variables for database configuration:
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`
- `AFL_SERVICE_URL`, `FFL_SERVICE_URL` (for gateway routing)
- `PORT` (for service/gateway port configuration)