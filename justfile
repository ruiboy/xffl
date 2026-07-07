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
    @echo "Postgres ready on :${DB_PORT:-5432} | Typesense ready on :${TYPESENSE_PORT:-8108} | Apollo Router ready on :${ROUTER_PORT:-4000}"

# Load test data into Postgres
dev-seed:
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/01_afl_seed.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/02_ffl_seed.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/03_afl_historical.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/04_ffl_players.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/05_ffl_round_data.sql
    docker exec -i xffl-postgres psql -U postgres -d xffl < dev/postgres/seed/06_seed_finalize.sql
    @echo "Test data loaded"

# Stop local infrastructure
dev-down:
    docker compose -f dev/docker-compose.yml down

# Stop infrastructure and delete all data
dev-reset:
    #!/usr/bin/env bash
    read -p "IMPORTANT!!! Do you need to back up the database first? [y/N] " answer
    if [[ "$answer" =~ ^[Yy]$ ]]; then
        echo "Aborting — run 'just db-backup' first, then re-run 'just dev-reset'."
        exit 1
    fi
    docker compose -f dev/docker-compose.yml down -v
    echo "Infrastructure stopped and volumes removed"

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

# Compose Apollo supergraph from local schema files
# Install rover: curl -sSL https://rover.apollo.dev/nix/latest | sh
# Router hot-reloads supergraph.graphql automatically via --hot-reload
supergraph-compose:
    cat services/afl/api/graphql/common.graphqls services/afl/api/graphql/query.graphqls services/afl/api/graphql/mutation.graphqls > dev/router/afl-schema.graphql
    cat services/ffl/api/graphql/common.graphqls services/ffl/api/graphql/query.graphqls services/ffl/api/graphql/mutation.graphqls > dev/router/ffl-schema.graphql
    ~/.rover/bin/rover supergraph compose --config dev/router/supergraph.yaml --elv2-license=accept --output dev/router/supergraph.graphql
    @echo "Supergraph written to dev/router/supergraph.graphql"

# Generate Twirp + protobuf Go code from contracts/proto/ → contracts/gen/
proto-gen:
    buf generate

# Export historical AFL stats from afltables.com to afl-historical/<season>.csv
# Walks backwards from `from` to `to` (inclusive). e.g. just afl-historical-export 2023 2023
afl-historical-export from to:
    cd services/afl && go run ./cmd/afltables-export -from {{from}} -to {{to}} -out ../../afl-historical

# Run AFL service tests (includes integration tests via testcontainers)
test-afl:
    cd services/afl && go test -tags integration ./...

# Run FFL service tests
test-ffl:
    cd services/ffl && go test -tags integration ./...

# Run frontend unit tests (vitest — pure logic only)
test-frontend:
    cd frontend/web && npm run test:unit

# Run e2e tests in a fully isolated environment (dev stack may remain running)
test-e2e:
    #!/usr/bin/env bash
    set -euo pipefail

    # Copy schemas into test-e2e (seed files live there permanently)
    cp dev/postgres/init/01_afl_schema.sql dev/postgres/test-e2e/01_afl_schema.sql
    cp dev/postgres/init/02_ffl_schema.sql dev/postgres/test-e2e/02_ffl_schema.sql
    cp dev/postgres/init/03_dataops.sql dev/postgres/test-e2e/03_dataops.sql

    docker compose -p xffl-test -f dev/docker-compose.test.yml up -d --force-recreate
    echo "Waiting for test Postgres on :5433..."
    until docker exec xffl-postgres-test pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
    echo "Test Postgres ready"
    echo "Waiting for test Router on :4001..."
    until curl -s http://localhost:4001/ >/dev/null 2>&1; do sleep 1; done
    echo "Test Router ready"

    # If npx (or npm run) is blocked on your machine, swap the command below for:
    #   ./node_modules/.bin/playwright test
    # (after running `npm ci` in frontend/web at least once to populate node_modules).
    (cd frontend/web && npx playwright test); STATUS=$?

    docker compose -p xffl-test -f dev/docker-compose.test.yml down
    rm -f dev/postgres/test-e2e/01_afl_schema.sql dev/postgres/test-e2e/02_ffl_schema.sql dev/postgres/test-e2e/03_dataops.sql
    exit $STATUS

# Run all tests (AFL unit, FFL unit, and e2e)
test-all:
    just test-afl
    just test-ffl
    just test-frontend
    just test-e2e

# Back up Postgres to backups/ (set BACKUP_REMOTE=rclone-remote:bucket/path to also upload)
db-backup:
    @bash dev/backup/backup.sh

# Restore Postgres from a backup file (defaults to latest in backups/)
db-restore file="":
    #!/usr/bin/env bash
    bash dev/backup/restore.sh {{file}}

# Snapshot AI control plane context for sharing or LLM ingestion
ai-snapshot:
    bash dev/ai-snapshot.sh

