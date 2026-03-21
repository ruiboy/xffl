-- name: FindMatchesByRoundID :many
SELECT id, round_id,
       COALESCE(home_club_match_id, 0) AS home_club_match_id,
       COALESCE(away_club_match_id, 0) AS away_club_match_id,
       COALESCE(venue, '') AS venue,
       COALESCE(start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(drv_result, '') AS drv_result
FROM afl.match
WHERE round_id = $1 AND deleted_at IS NULL;

-- name: FindMatchByID :one
SELECT id, round_id,
       COALESCE(home_club_match_id, 0) AS home_club_match_id,
       COALESCE(away_club_match_id, 0) AS away_club_match_id,
       COALESCE(venue, '') AS venue,
       COALESCE(start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(drv_result, '') AS drv_result
FROM afl.match
WHERE id = $1 AND deleted_at IS NULL;
