-- name: FindMatchesByRoundID :many
SELECT m.id, m.round_id,
       COALESCE(m.match_style, '') AS match_style,
       COALESCE(home.id, 0) AS home_club_match_id,
       COALESCE(away.id, 0) AS away_club_match_id,
       COALESCE(m.venue, '') AS venue,
       COALESCE(m.start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(m.drv_result, '') AS drv_result
FROM ffl.match m
LEFT JOIN ffl.club_match home ON home.match_id = m.id AND home.side = 'home' AND home.deleted_at IS NULL
LEFT JOIN ffl.club_match away ON away.match_id = m.id AND away.side = 'away' AND away.deleted_at IS NULL
WHERE m.round_id = $1 AND m.deleted_at IS NULL
ORDER BY m.id;

-- name: FindMatchByID :one
SELECT m.id, m.round_id,
       COALESCE(m.match_style, '') AS match_style,
       COALESCE(home.id, 0) AS home_club_match_id,
       COALESCE(away.id, 0) AS away_club_match_id,
       COALESCE(m.venue, '') AS venue,
       COALESCE(m.start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(m.drv_result, '') AS drv_result
FROM ffl.match m
LEFT JOIN ffl.club_match home ON home.match_id = m.id AND home.side = 'home' AND home.deleted_at IS NULL
LEFT JOIN ffl.club_match away ON away.match_id = m.id AND away.side = 'away' AND away.deleted_at IS NULL
WHERE m.id = $1 AND m.deleted_at IS NULL;

-- name: FindMatchesByIDs :many
SELECT m.id, m.round_id,
       COALESCE(m.match_style, '') AS match_style,
       COALESCE(home.id, 0) AS home_club_match_id,
       COALESCE(away.id, 0) AS away_club_match_id,
       COALESCE(m.venue, '') AS venue,
       COALESCE(m.start_dt, '0001-01-01T00:00:00Z'::timestamptz) AS start_dt,
       COALESCE(m.drv_result, '') AS drv_result
FROM ffl.match m
LEFT JOIN ffl.club_match home ON home.match_id = m.id AND home.side = 'home' AND home.deleted_at IS NULL
LEFT JOIN ffl.club_match away ON away.match_id = m.id AND away.side = 'away' AND away.deleted_at IS NULL
WHERE m.id = ANY(@ids::int[]) AND m.deleted_at IS NULL
ORDER BY m.id;

-- name: UpdateFflMatchResult :exec
UPDATE ffl.match
SET drv_result = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateMatch :one
INSERT INTO ffl.match (round_id, match_style)
VALUES ($1, sqlc.narg('match_style'))
RETURNING id, round_id;

-- Matches are removed outright rather than soft-deleted: the fixture builder
-- only edits rounds with no submitted teams, so a dropped match holds nothing
-- worth keeping, and tombstones would accumulate on every save. club_match rows
-- follow via ON DELETE CASCADE.

-- name: DeleteMatchesByRoundID :exec
DELETE FROM ffl.match WHERE round_id = $1;

-- name: DeleteMatchByID :exec
DELETE FROM ffl.match WHERE id = $1;
