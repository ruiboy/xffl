set dotenv-load := true

# List available recipes
default:
    @just --list

# Start local infrastructure (Postgres + Zinc)
dev-up:
    docker compose -f dev/docker-compose.yml up -d
    @echo "Waiting for Postgres..."
    @until docker exec xffl-postgres pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
    @echo "Postgres ready on :${DB_PORT:-5432} | Zinc ready on :${ZINC_PORT:-4080}"

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

# Snapshot AI control plane context for sharing or LLM ingestion
ai-snapshot:
    bash dev/ai-snapshot.sh

# Run AFL service (port 8080)
run-afl:
    cd services/afl && go run ./cmd/main.go

# Run FFL service (port 8081)
run-ffl:
    cd services/ffl && go run ./cmd/main.go

# Run Search service (port 8082)
run-search:
    cd services/search && go run ./cmd/main.go

# Run Gateway (port 8090)
run-gateway:
    cd services/gateway && go run ./cmd/main.go

# Run Frontend (port 3000)
run-frontend:
    cd frontend/web && npm run dev

# Run AFL + FFL services, gateway, and frontend together
run-all:
    #!/usr/bin/env bash
    trap 'kill 0' EXIT
    just run-afl &
    just run-ffl &
    just run-gateway &
    just run-frontend &
    wait

# Stop all running services (AFL, FFL, gateway, frontend)
stop-all:
    #!/usr/bin/env bash
    for port in 8080 8081 8090 3000; do
        pid=$(lsof -ti :$port 2>/dev/null)
        if [ -n "$pid" ]; then
            kill $pid 2>/dev/null && echo "Stopped process on :$port (PID $pid)"
        fi
    done

# Run AFL service tests
test-afl:
    cd services/afl && go test ./...

# Run FFL service tests
test-ffl:
    cd services/ffl && go test ./...

# Run e2e tests in a fully isolated environment (dev stack may remain running)
test-e2e:
    #!/usr/bin/env bash
    set -euo pipefail

    # Assemble test-init from canonical sources (single source of truth)
    mkdir -p dev/postgres/test-init
    cp dev/postgres/init/01_afl_schema.sql dev/postgres/test-init/01_afl_schema.sql
    cp dev/postgres/init/02_ffl_schema.sql dev/postgres/test-init/02_ffl_schema.sql
    cp dev/postgres/seed/01_afl_seed.sql   dev/postgres/test-init/03_afl_seed.sql
    cp dev/postgres/seed/02_ffl_seed.sql   dev/postgres/test-init/04_ffl_seed.sql

    docker compose -f dev/docker-compose.test.yml up -d
    echo "Waiting for test Postgres on :5433..."
    until docker exec xffl-postgres-test pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
    echo "Test Postgres ready"

    cd frontend/web && npx playwright test; STATUS=$?

    docker compose -f dev/docker-compose.test.yml down
    rm -rf dev/postgres/test-init
    exit $STATUS

