-- name: FindByesByRoundID :many
SELECT id, round_id, club_season_id
FROM afl.bye
WHERE round_id = $1 AND deleted_at IS NULL;

-- name: FindByeByRoundAndClub :one
SELECT id, round_id, club_season_id
FROM afl.bye
WHERE round_id = $1 AND club_season_id = $2 AND deleted_at IS NULL;

-- name: FindByesByRoundIDWithClub :many
SELECT b.id, b.round_id, b.club_season_id, c.id AS club_id, c.name AS club_name
FROM afl.bye b
JOIN afl.club_season cs ON cs.id = b.club_season_id AND cs.deleted_at IS NULL
JOIN afl.club c ON c.id = cs.club_id AND c.deleted_at IS NULL
WHERE b.round_id = $1 AND b.deleted_at IS NULL
ORDER BY c.name;

-- name: UpsertBye :one
INSERT INTO afl.bye (round_id, club_season_id)
VALUES ($1, $2)
ON CONFLICT (round_id, club_season_id)
DO UPDATE SET updated_at = CURRENT_TIMESTAMP
WHERE afl.bye.deleted_at IS NULL
RETURNING id, round_id, club_season_id;
