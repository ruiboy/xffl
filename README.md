# GFFL (GraphQL Fantasy Football League)

A full-stack web application for managing a fantasy football league with GraphQL API.

## Project Structure

```
gffl/
├── backend/           # Go backend with GraphQL API
│   ├── graph/        # GraphQL schema and resolvers
│   ├── main.go       # Application entry point and server setup
│   └── gqlgen.yml    # GraphQL code generation config
└── frontend/         # Vue.js frontend
    ├── src/          # Source code
    ├── public/       # Static assets
    └── index.html    # Entry HTML file
```

## Prerequisites

- Go 1.16 or later
- Node.js 16 or later
- PostgreSQL 13 or later
- npm or yarn

## Backend Setup

The backend is built with Go and uses gqlgen for GraphQL API generation.

### Key Components

- `main.go`: The application entry point that:
  - Sets up the GraphQL server
  - Configures CORS for frontend communication
  - Handles HTTP routing
  - Sets up the GraphQL playground
  - Manages server configuration and environment variables

### GraphQL Code Generation

After making changes to your GraphQL schema or resolvers, you'll need to regenerate the code:

```bash
cd backend
go run github.com/99designs/gqlgen generate
```

This will:
- Generate type-safe Go code from your GraphQL schema
- Update resolver interfaces
- Create new resolver stubs for any new queries or mutations

### Running the Backend

```bash
cd backend
go run main.go
```

The GraphQL server will start on `http://localhost:8080` with the following endpoints:
- `/query` - GraphQL API endpoint
- `/` - GraphQL playground for testing queries

## Frontend Setup

The frontend is built with Vue 3, TypeScript, and Vite.

### Key Components

- `src/App.vue` - Root component
- `src/router/` - Vue Router configuration
- `src/views/` - Page components
- `src/components/` - Reusable components
- `src/stores/` - Pinia state management
- `src/graphql/` - GraphQL queries and mutations

### Running the Frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at `http://localhost:3000`.

## Development

1. Start the backend server:
   ```bash
   cd backend
   go run main.go
   ```

2. Start the frontend development server:
   ```bash
   cd frontend
   npm run dev
   ```

3. Access the application:
   - Frontend: http://localhost:3000
   - GraphQL Playground: http://localhost:8080

## GraphQL API

The backend provides a GraphQL API with the following features:
- Query and mutation support
- CORS configuration for frontend access
- GraphQL playground for testing
- Automatic persisted queries
- Query caching

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
