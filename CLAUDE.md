# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Backend (Go)
- Start backend server: `cd backend && go run main.go`
- Generate GraphQL code: `cd backend && go run github.com/99designs/gqlgen generate`
- Build: `cd backend && go build`

### Frontend (Vue.js)
- Install dependencies: `cd frontend && npm install`
- Start dev server: `cd frontend && npm run dev`
- Build for production: `cd frontend && npm run build && npm run preview`

### Database
- Run migrations: `psql -U postgres -d gffl -f backend/db/migrations/001_create_ffl_tables_up.sql`
- Insert test data: `psql -U postgres -d gffl -f backend/db/test_scripts/insert_test_clubs.sql`

## Architecture

This is a full-stack fantasy football league application with:

- **Backend**: Go with GraphQL API using gqlgen, GORM for database ORM, PostgreSQL database
- **Frontend**: Vue 3 + TypeScript + Vite, Apollo Client for GraphQL, PrimeVue UI components
- **Database Schema**: Uses `ffl` schema with `club` and `player` tables, managed via SQL migrations

### Key Backend Components

- `main.go`: Server setup with GraphQL handler, CORS configuration, and playground
- `graph/schema.graphqls`: GraphQL schema definitions
- `graph/schema.resolvers.go`: Resolver implementations (auto-generated stubs)
- `db/db.go`: Database models (FFLClub, FFLPlayer) and connection initialization
- Database uses environment variables (DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT)

### Key Frontend Components

- Uses Apollo Client for GraphQL queries/mutations
- Vue Router for navigation between Home and Players views
- PrimeVue for UI components
- State management with Pinia

### Development Workflow

1. After modifying GraphQL schema, always run `cd backend && go run github.com/99designs/gqlgen generate`
2. Backend runs on :8080 with GraphQL playground at root and API at /query
3. Frontend runs on :3000 with CORS configured for localhost communication
4. Database changes require manual SQL migration execution