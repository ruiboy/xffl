-- name: UpsertMatchSourceMap :exec
INSERT INTO afl.match_source_map (source, external_id, match_id, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
ON CONFLICT (source, match_id) DO UPDATE
    SET external_id = EXCLUDED.external_id,
        updated_at  = CURRENT_TIMESTAMP;

-- name: FindMatchSourceMapByMatchID :one
SELECT source, external_id, match_id
FROM afl.match_source_map
WHERE source = $1 AND match_id = $2;
