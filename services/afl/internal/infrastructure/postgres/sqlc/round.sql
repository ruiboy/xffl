-- name: FindRoundsBySeasonID :many
SELECT id, name, season_id
FROM afl.round
WHERE season_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: FindRoundByID :one
SELECT id, name, season_id
FROM afl.round
WHERE id = $1 AND deleted_at IS NULL;
