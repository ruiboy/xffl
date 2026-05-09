-- name: FindMatchesByRoundID :many
SELECT id, round_id,
       COALESCE(home_club_match_id, 0) AS home_club_match_id,
       COALESCE(away_club_match_id, 0) AS away_club_match_id,
       COALESCE(venue, '') AS venue,
       COALESCE(start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(drv_result, '') AS drv_result
FROM ffl.match
WHERE round_id = $1 AND deleted_at IS NULL;

-- name: FindMatchByID :one
SELECT id, round_id,
       COALESCE(home_club_match_id, 0) AS home_club_match_id,
       COALESCE(away_club_match_id, 0) AS away_club_match_id,
       COALESCE(venue, '') AS venue,
       COALESCE(start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(drv_result, '') AS drv_result
FROM ffl.match
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindMatchesByIDs :many
SELECT id, round_id,
       COALESCE(home_club_match_id, 0) AS home_club_match_id,
       COALESCE(away_club_match_id, 0) AS away_club_match_id,
       COALESCE(venue, '') AS venue,
       COALESCE(start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(drv_result, '') AS drv_result
FROM ffl.match
WHERE id = ANY(@ids::int[]) AND deleted_at IS NULL;

-- name: UpdateFflMatchResult :exec
UPDATE ffl.match
SET drv_result = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindFinalFflMatchesBySeasonID :many
SELECT m.id, m.round_id,
       home.id             AS home_club_match_id,
       home.club_season_id AS home_club_season_id,
       COALESCE(home.drv_score, 0) AS home_score,
       away.id             AS away_club_match_id,
       away.club_season_id AS away_club_season_id,
       COALESCE(away.drv_score, 0) AS away_score
FROM ffl.match m
JOIN ffl.round r ON r.id = m.round_id
JOIN ffl.club_match home ON home.id = m.home_club_match_id
     AND home.data_status = 'final' AND home.deleted_at IS NULL
JOIN ffl.club_match away ON away.id = m.away_club_match_id
     AND away.data_status = 'final' AND away.deleted_at IS NULL
WHERE r.season_id = $1 AND m.deleted_at IS NULL;
