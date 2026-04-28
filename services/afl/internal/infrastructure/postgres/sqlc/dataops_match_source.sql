-- name: UpsertDataopsMatchSource :exec
INSERT INTO afl.dataops_match_source (source, external_id, match_id, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
ON CONFLICT (source, match_id) DO UPDATE
    SET external_id = EXCLUDED.external_id,
        updated_at  = CURRENT_TIMESTAMP;

-- name: FindDataopsMatchSourceByMatchID :one
SELECT source, external_id, match_id
FROM afl.dataops_match_source
WHERE source = $1 AND match_id = $2;
