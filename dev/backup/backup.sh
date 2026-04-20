#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BACKUP_DIR="${REPO_ROOT}/backups"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
FILENAME="postgres_${TIMESTAMP}.sql.gz"
FILEPATH="${BACKUP_DIR}/${FILENAME}"

mkdir -p "${BACKUP_DIR}"

echo "Backing up xffl postgres → ${FILEPATH}"
docker exec xffl-postgres pg_dump -U postgres --data-only --disable-triggers xffl | gzip > "${FILEPATH}"

SIZE="$(du -sh "${FILEPATH}" | cut -f1)"
echo "Done: ${FILENAME} (${SIZE})"

if [ -n "${BACKUP_REMOTE:-}" ]; then
    echo "Uploading to ${BACKUP_REMOTE} ..."
    rclone copy "${FILEPATH}" "${BACKUP_REMOTE}"
    echo "Upload complete"
else
    echo "Skipping upload (BACKUP_REMOTE not set)"
fi