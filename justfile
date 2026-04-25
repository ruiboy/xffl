set dotenv-load := true

log_level := env_var_or_default("LOG_LEVEL", "debug")

# List available recipes
default:
    @just --list

# Start local infrastructure (Postgres + Typesense)
dev-up:
    docker compose -f dev/docker-compose.yml up -d
    @echo "Waiting for Postgres..."
    @until docker exec xffl-postgres pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
    @echo "Postgres ready on :${DB_PORT:-5432} | Typesense ready on :${TYPESENSE_PORT:-8108}"

# Load test data into Postgres
dev-seed:
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/01_afl_seed.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/02_ffl_seed.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/03_afl_historical.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/04_ffl_players.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/05_ffl_trades.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/06_ffl_round_data.sql
    @echo "Test data loaded"

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

# Run AFL service (port 8080)
run-afl:
    cd services/afl && LOG_LEVEL={{log_level}} go run ./cmd/main.go

# Run FFL service (port 8081)
run-ffl:
    cd services/ffl && LOG_LEVEL={{log_level}} go run ./cmd/main.go

# Run Search service (port 8082)
run-search:
    cd services/search && LOG_LEVEL={{log_level}} go run ./cmd/main.go

# Run Gateway (port 8090)
run-gateway:
    cd services/gateway && LOG_LEVEL={{log_level}} go run ./cmd/main.go

# Install frontend dependencies (run once before first run-all)
install-frontend:
    cd frontend/web && npm install

# Run Frontend (port 3000)
run-frontend: install-frontend
    cd frontend/web && npm run dev

# Run AFL + FFL + Search services, gateway, and frontend together
run-all:
    #!/usr/bin/env bash
    trap 'kill 0' EXIT
    just run-afl &
    just run-ffl &
    just run-search &
    just run-gateway &
    just run-frontend &
    wait

# Stop all running services (AFL, FFL, Search, gateway, frontend)
stop-all:
    #!/usr/bin/env bash
    for port in 8080 8081 8082 8090 3000; do
        pid=$(lsof -ti :$port 2>/dev/null)
        if [ -n "$pid" ]; then
            kill $pid 2>/dev/null && echo "Stopped process on :$port (PID $pid)"
        fi
    done

# Run AFL service tests (includes integration tests via testcontainers)
test-afl:
    cd services/afl && go test -tags integration ./...

# Run FFL service tests
test-ffl:
    cd services/ffl && go test ./...

# Run e2e tests in a fully isolated environment (dev stack may remain running)
test-e2e:
    #!/usr/bin/env bash
    set -euo pipefail

    # Copy schemas into test-e2e (seed files live there permanently)
    cp dev/postgres/init/01_afl_schema.sql dev/postgres/test-e2e/01_afl_schema.sql
    cp dev/postgres/init/02_ffl_schema.sql dev/postgres/test-e2e/02_ffl_schema.sql

    docker compose -p xffl-test -f dev/docker-compose.test.yml up -d --force-recreate
    echo "Waiting for test Postgres on :5433..."
    until docker exec xffl-postgres-test pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
    echo "Test Postgres ready"

    (cd frontend/web && npx playwright test); STATUS=$?

    docker compose -p xffl-test -f dev/docker-compose.test.yml down
    rm -f dev/postgres/test-e2e/01_afl_schema.sql dev/postgres/test-e2e/02_ffl_schema.sql
    exit $STATUS

# Run all tests (AFL unit, FFL unit, and e2e)
test-all:
    just test-afl
    just test-ffl
    just test-e2e

# Back up Postgres to backups/ (set BACKUP_REMOTE=rclone-remote:bucket/path to also upload)
backup-db:
    @bash dev/backup/backup.sh

# Restore Postgres from a backup file (defaults to latest in backups/)
restore-db file="":
    #!/usr/bin/env bash
    bash dev/backup/restore.sh {{file}}

# Snapshot AI control plane context for sharing or LLM ingestion
ai-snapshot:
    bash dev/ai-snapshot.sh

