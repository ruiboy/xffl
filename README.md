# xffl 

FFL as a modular monolith, with experimentation into DDD, bounded contexts, clean architecture, CQRS, search.
Totally over engineered for what it does, but, experimenting.

(FFL = Fantasy Football League)

Primary techs are golang, graphql, postgres, Vue, Zinc

Built with a lot of code agent.

## Prerequisites

- Go 1.16 or later
- Node.js 16 or later
- PostgreSQL 13 or later
- npm or yarn

## Database Setup

The application uses PostgreSQL as its database. Here's how to set it up:

### Installing PostgreSQL

#### macOS
```bash
# Install PostgreSQL using Homebrew
brew install postgresql@14

# Start PostgreSQL as service
# If using port other than 5432, you may need to adjust configuration, eg at
# /opt/homebrew/var/postgresql@14/postgresql.conf
brew services start postgresql@14

# Or, if you don't want/need a background service you can just run:
/usr/local/opt/postgresql@14/bin/postgres -D /usr/local/var/postgresql@14

# Create the database
createdb xffl

# Create a PostgreSQL user (if not exists)
createuser -s postgres

# Set password for postgres user
psql postgres -c "ALTER USER postgres WITH PASSWORD 'postgres';"
```

#### Linux (Ubuntu/Debian)
```bash
# Install PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib

# Start PostgreSQL service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create the database
sudo -u postgres createdb xffl

# Set password for postgres user
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD 'postgres';"
```

### Database Configuration

The application uses environment variables for database configuration. Create a `.env` file in the backend directory with the following default settings:

```
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=xffl
DB_PORT=5423
```

You can modify these values in the `.env` file to match your database setup.

### Running Migrations

The database schema is managed through SQL migration files. To apply the migrations:

```bash
psql -U postgres -d xffl -f backend/internal/adapters/persistence/migrations/001_create_afl_tables_up.sql
psql -U postgres -d xffl -f backend/internal/adapters/persistence/migrations/002_create_ffl_tables_up.sql
```

You can revert by running the down migrations.

Alternatively you can run interactively:

```bash
# Connect to the database
psql -U postgres -d xffl

# Run the migration file
\i backend/...sql
```

### Test Data

Test data scripts are available in the `backend/db/test_scripts` directory. To insert test data:

```bash
psql -U postgres -d xffl -f backend/internal/adapters/persistence/test_scripts/insert_ffl_data.sql
```

### Verifying the Connection

To verify that the database connection is working:

1. Start the backend server:
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. Check the logs for any database connection errors. If there are no errors, the connection is successful.

## Backend Setup

The backend is built with Go and uses gqlgen for GraphQL API generation.

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
go run cmd/server/main.go
```

The GraphQL server will start on `http://localhost:8080` with the following endpoints:
- `/query` - GraphQL API endpoint
- `/` - GraphQL playground for testing queries

## Web 
Frontend Setup

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
   go run cmd/server/main.go
   ```

2. Start the frontend development server:
   ```bash
   cd frontend
   npm run dev
   ```

3. Access the application:
   - Frontend: http://localhost:3000
   - GraphQL Playground: http://localhost:8080

## Project Structure

```
xffl/
├── api/                     # API definitions (easily discoverable)
│   └── graphql/            # GraphQL schema definitions
│       └── schema.graphqls
├── backend/           # Go backend with Clean + Hexagonal Architecture
│   ├── cmd/
│   │   └── server/   # Application entry point and server setup
│   ├── internal/
│   │   ├── domain/          # Pure domain: entities, value objects, domain events
│   │   ├── application/     # Use cases, application services
│   │   ├── adapters/
│   │   │   ├── graphql/     # GraphQL resolvers (input adapters)
│   │   │   ├── rest/        # REST controllers (optional)
│   │   │   ├── persistence/ # DB implementations (output adapters)
│   │   │   └── pubsub/      # Event publishing (e.g., Kafka, NATS)
│   │   ├── ports/
│   │   │   ├── in/          # Interfaces for input (resolver/service interfaces)
│   │   │   └── out/         # Interfaces for output (repo, event bus)
│   │   └── infrastructure/  # DB config, HTTP server setup, etc.
│   ├── pkg/
│   │   └── shared/          # Common helpers, errors, utils
│   ├── go.mod
│   ├── go.sum
│   └── gqlgen.yml          # GraphQL code generation config
└── frontend/         # Vue.js frontend
    ├── src/          # Source code
    ├── public/       # Static assets
    └── index.html    # Entry HTML file
```

## Architecture

### Frontend

#### Key Components

- `src/App.vue` - Root component
- `src/router/` - Vue Router configuration
- `src/views/` - Page components
- `src/components/` - Reusable components
- `src/stores/` - Pinia state management
- `src/graphql/` - GraphQL queries and mutations

### Backend

The backend follows Clean Architecture + Hexagonal Architecture principles:

#### Clean Architecture Layers:
- **Domain Layer** (`internal/domain/`): Pure business logic, entities, value objects, domain events
- **Application Layer** (`internal/application/`): Use cases and application services that orchestrate domain operations
- **Interface Adapters** (`internal/adapters/`): Adapters for external systems (GraphQL, REST, Database, PubSub)
- **Infrastructure** (`internal/infrastructure/`): Framework and tools configuration

#### Hexagonal Architecture Ports:
- **Input Ports** (`internal/ports/in/`): Interfaces defining how external systems can interact with the application
- **Output Ports** (`internal/ports/out/`): Interfaces defining how the application interacts with external systems

#### Key Components:
- `cmd/server/main.go`: Application entry point and server setup
- `internal/adapters/graphql/`: GraphQL resolvers (input adapters)
- `internal/adapters/persistence/`: Database implementations (output adapters)
- `internal/domain/`: Business entities and domain logic
- `internal/application/`: Use cases and application services
- `internal/infrastructure/`: Database connections, HTTP server configuration

### Data Model

The application is designed to manage two types of leagues: AFL (Australian Football League) and FFL (Fantasy Football League). Each league has its own set of entities, including seasons, rounds, matches, clubs, and players.

Ostensibly, data is stored in first normal form (1NF). However, at this stage, to optimize read performance for the frontend, some data is denormalized. This includes pre-calculated fields like scores, premiership points, and match results.

#### AFL League Data Model

```plantuml 
@startuml
hide empty members

skinparam class {
  BackgroundColor<<league>> Gold
  BackgroundColor<<club>> LightBlue 
  BackgroundColor<<player>> LightGreen
}

class League <<league>> {
  eg. AFL
}
class Season <<league>> {
  eg. AFL 2025 
}
class Round <<league>> {
  eg. AFL 2025 Rd 1
}
class Match <<league>> {
  eg. CROWS v NM in AFL 2025 Rd 1
  home_club_match_id
  away_club_match_id
  venue
  start_dt
  --
  <<denormalized>>
  result: home_win|away_win|draw|no_result
}

class Club <<club>> {
  eg. CROWS
}
class ClubSeason <<club>> {
  eg. CROWS in AFL 2025
  --
  <<denormalized>>
  played
  won
  lost
  drawn
  for
  against
  premiership_points
}
class ClubMatch <<club>> {
  eg. CROWS in CROWS v NM in AFL 2025 Rd 1
  rushed_behinds
  --
  <<denormalized>>
  score
  premiership_points
}

class Player <<player>> {
  eg. Dawson
}
class PlayerSeason <<player>> {
  eg. Dawson in CROWS in AFL 2025
}
class PlayerMatch <<player>> {
  eg. Dawson in CROWS in CROWS v NM in AFL 2025 Rd 1
  kicks
  handballs
  marks
  hitouts
  tackles
  goals
  behinds
}

League *-- "0..*" Season
Season *-- "0..*" Round
Round *-- "0..*" Match

Club *-- "0..*" ClubSeason
Season *-- "0..*" ClubSeason

Match *-- "2" ClubMatch
ClubSeason *-- "0..*" ClubMatch

Player *-- "0..*" PlayerSeason
ClubSeason *-- "0..*" PlayerSeason

ClubMatch *-- "0..*" PlayerMatch
PlayerSeason *-- "0..*" PlayerMatch

@enduml
```

![AFL ERD](doc/erd-afl.png)

#### FFL League Data Model

```plantuml 
@startuml
hide empty members

skinparam class {
  BackgroundColor<<league>> Gold
  BackgroundColor<<club>> LightBlue 
  BackgroundColor<<player>> LightGreen
}

class League <<league>> {
  eg. FFL
}
class Season <<league>> {
  eg. FFL 2025 
}
class Round <<league>> {
  eg. FFL 2025 Rd 1
}
class Match <<league>> {
  eg. ROOS v FRED in FFL 2025 Rd 1
  match_style: versus|bye|super_bye
  clubs[]
  --
  <<denormalized>>
  result
}

class Club <<club>> {
  eg. ROOS
}
class ClubSeason <<club>> {
  eg. ROOS in FFL 2025
  --
  <<denormalized>>
  played
  won
  lost
  drawn
  for
  against
  extra_points
  premiership_points
}
class ClubMatch <<club>> {
  eg. ROOS in ROOS v FRED in FFL 2025 Rd 1
  --
  <<denormalized>>
  score
  premiership_points
}

class Player <<player>> {
  eg. Dawson
}
class PlayerSeason <<player>> {
  eg. Dawson in ROOS in FFL 2025
  from_round_id
  to_round_id
}
class PlayerMatch <<player>> {
  eg. Dawson in ROOS in ROOS v FRED in FFL 2025 Rd 1
  position
  interchange_positions
  status: dnp|subbed_in
  score
}

League *-- "0..*" Season
Season *-- "0..*" Round
Round *-- "0..*" Match

Club *-- "0..*" ClubSeason
Season *-- "0..*" ClubSeason

Match *-- "2" ClubMatch
ClubSeason *-- "0..*" ClubMatch

Player *-- "0..*" PlayerSeason
ClubSeason *-- "0..*" PlayerSeason

ClubMatch *-- "0..*" PlayerMatch
PlayerSeason *-- "0..*" PlayerMatch

@enduml
```

![FFL ERD](doc/erd-ffl.png)
