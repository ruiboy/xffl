-- name: FindMatchesByRoundID :many
SELECT m.id, m.round_id,
       COALESCE(home.id, 0) AS home_club_match_id,
       COALESCE(away.id, 0) AS away_club_match_id,
       COALESCE(m.venue, '') AS venue,
       COALESCE(m.start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(m.drv_result, '') AS drv_result,
       m.data_status
FROM afl.match m
LEFT JOIN afl.club_match home ON home.match_id = m.id AND home.side = 'home' AND home.deleted_at IS NULL
LEFT JOIN afl.club_match away ON away.match_id = m.id AND away.side = 'away' AND away.deleted_at IS NULL
WHERE m.round_id = $1 AND m.deleted_at IS NULL;

-- name: FindMatchByID :one
SELECT m.id, m.round_id,
       COALESCE(home.id, 0) AS home_club_match_id,
       COALESCE(away.id, 0) AS away_club_match_id,
       COALESCE(m.venue, '') AS venue,
       COALESCE(m.start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(m.drv_result, '') AS drv_result,
       m.data_status
FROM afl.match m
LEFT JOIN afl.club_match home ON home.match_id = m.id AND home.side = 'home' AND home.deleted_at IS NULL
LEFT JOIN afl.club_match away ON away.match_id = m.id AND away.side = 'away' AND away.deleted_at IS NULL
WHERE m.id = $1 AND m.deleted_at IS NULL;

-- name: FindMatchesByIDs :many
SELECT m.id, m.round_id,
       COALESCE(home.id, 0) AS home_club_match_id,
       COALESCE(away.id, 0) AS away_club_match_id,
       COALESCE(m.venue, '') AS venue,
       COALESCE(m.start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(m.drv_result, '') AS drv_result,
       m.data_status
FROM afl.match m
LEFT JOIN afl.club_match home ON home.match_id = m.id AND home.side = 'home' AND home.deleted_at IS NULL
LEFT JOIN afl.club_match away ON away.match_id = m.id AND away.side = 'away' AND away.deleted_at IS NULL
WHERE m.id = ANY(@ids::int[]) AND m.deleted_at IS NULL;

-- name: UpdateMatchDataStatus :exec
UPDATE afl.match
SET data_status = $2,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateMatchResult :exec
UPDATE afl.match
SET drv_result = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindFinalMatchesBySeasonID :many
SELECT m.id,
       m.round_id,
       COALESCE(home.id, 0)              AS home_club_match_id,
       COALESCE(home.club_season_id, 0)  AS home_club_season_id,
       COALESCE(home.drv_score, 0)       AS home_score,
       COALESCE(away.id, 0)              AS away_club_match_id,
       COALESCE(away.club_season_id, 0)  AS away_club_season_id,
       COALESCE(away.drv_score, 0)       AS away_score
FROM afl.match m
JOIN afl.round r ON r.id = m.round_id AND r.deleted_at IS NULL
LEFT JOIN afl.club_match home ON home.match_id = m.id AND home.side = 'home' AND home.deleted_at IS NULL
LEFT JOIN afl.club_match away ON away.match_id = m.id AND away.side = 'away' AND away.deleted_at IS NULL
WHERE r.season_id = $1
  AND m.data_status = 'final'
  AND m.deleted_at IS NULL;
