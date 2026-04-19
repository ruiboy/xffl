#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BACKUP_DIR="${REPO_ROOT}/backups"

# Use provided file or latest backup
if [ -n "${1:-}" ]; then
    FILEPATH="$1"
else
    FILEPATH="$(ls -t "${BACKUP_DIR}"/postgres_*.sql.gz 2>/dev/null | head -1)"
    if [ -z "${FILEPATH}" ]; then
        echo "Error: no backup file found in ${BACKUP_DIR}" >&2
        exit 1
    fi
fi

if [ ! -f "${FILEPATH}" ]; then
    echo "Error: file not found: ${FILEPATH}" >&2
    exit 1
fi

echo "Restoring from: ${FILEPATH}"
echo "This will wipe and recreate the dev DB. Press Ctrl-C to cancel, Enter to continue."
read -r

echo "Resetting dev infrastructure..."
just -f "${REPO_ROOT}/justfile" dev-reset
just -f "${REPO_ROOT}/justfile" dev-up

echo "Waiting for Postgres to finish initialising..."
until docker exec xffl-postgres psql -U postgres -d xffl -c "SELECT 1" >/dev/null 2>&1; do sleep 1; done

echo "Restoring data..."
gunzip -c "${FILEPATH}" | docker exec -i xffl-postgres psql -U postgres -d xffl -v ON_ERROR_STOP=1 -q --tuples-only

echo "Restore complete. Verifying row counts..."
docker exec xffl-postgres psql -U postgres -d xffl -c "
SELECT 'afl.player'        AS \"table\", COUNT(*) FROM afl.player
UNION ALL
SELECT 'afl.player_match'  AS \"table\", COUNT(*) FROM afl.player_match
UNION ALL
SELECT 'afl.match'         AS \"table\", COUNT(*) FROM afl.match
UNION ALL
SELECT 'ffl.player'        AS \"table\", COUNT(*) FROM ffl.player
ORDER BY 1;
"