-- name: FindAllSeasons :many
SELECT id, name, league_id
FROM afl.season
WHERE deleted_at IS NULL
ORDER BY name;

-- name: FindSeasonByID :one
SELECT id, name, league_id
FROM afl.season
WHERE id = $1 AND deleted_at IS NULL;
