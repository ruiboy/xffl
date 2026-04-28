-- name: FindMatchesByRoundID :many
SELECT id, round_id,
       COALESCE(home_club_match_id, 0) AS home_club_match_id,
       COALESCE(away_club_match_id, 0) AS away_club_match_id,
       COALESCE(venue, '') AS venue,
       COALESCE(start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(drv_result, '') AS drv_result,
       stats_import_status,
       stats_imported_at
FROM afl.match
WHERE round_id = $1 AND deleted_at IS NULL;

-- name: FindMatchByID :one
SELECT id, round_id,
       COALESCE(home_club_match_id, 0) AS home_club_match_id,
       COALESCE(away_club_match_id, 0) AS away_club_match_id,
       COALESCE(venue, '') AS venue,
       COALESCE(start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(drv_result, '') AS drv_result,
       stats_import_status,
       stats_imported_at
FROM afl.match
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindMatchesByIDs :many
SELECT id, round_id,
       COALESCE(home_club_match_id, 0) AS home_club_match_id,
       COALESCE(away_club_match_id, 0) AS away_club_match_id,
       COALESCE(venue, '') AS venue,
       COALESCE(start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(drv_result, '') AS drv_result,
       stats_import_status,
       stats_imported_at
FROM afl.match
WHERE id = ANY(@ids::int[]) AND deleted_at IS NULL;

-- name: UpdateMatchImportStatus :exec
UPDATE afl.match
SET stats_import_status = $2,
    stats_imported_at   = $3,
    updated_at          = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
