set dotenv-load := true

# List available recipes
default:
    @just --list

# Start local infrastructure (Postgres + Zinc)
dev-up:
    docker compose -f dev/docker-compose.yml up -d
    @echo "Waiting for Postgres..."
    @until docker exec xffl-postgres pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
    @echo "Postgres ready on :5432 | Zinc ready on :4080"

# Stop local infrastructure
dev-down:
    docker compose -f dev/docker-compose.yml down

# Stop infrastructure and delete all data
dev-reset:
    docker compose -f dev/docker-compose.yml down -v
    @echo "Infrastructure stopped and volumes removed"

# Tail infrastructure logs
dev-logs:
    docker compose -f dev/docker-compose.yml logs -f

# Load test data into Postgres
dev-seed:
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/01_afl_seed.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/02_ffl_seed.sql
    @echo "Test data loaded"
