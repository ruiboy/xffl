-- name: FindByesByRoundID :many
SELECT id, round_id, club_season_id
FROM afl.bye
WHERE round_id = $1 AND deleted_at IS NULL;

-- name: FindByeByRoundAndClub :one
SELECT id, round_id, club_season_id
FROM afl.bye
WHERE round_id = $1 AND club_season_id = $2 AND deleted_at IS NULL;

-- name: UpsertBye :one
INSERT INTO afl.bye (round_id, club_season_id)
VALUES ($1, $2)
ON CONFLICT (round_id, club_season_id)
DO UPDATE SET updated_at = CURRENT_TIMESTAMP
WHERE afl.bye.deleted_at IS NULL
RETURNING id, round_id, club_season_id;
