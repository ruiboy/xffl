-- name: FindAllPlayers :many
SELECT id, name, afl_player_id
FROM ffl.player
WHERE deleted_at IS NULL
ORDER BY name;

-- name: FindPlayerByID :one
SELECT id, name, afl_player_id
FROM ffl.player
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePlayer :one
INSERT INTO ffl.player (name, afl_player_id)
VALUES ($1, $2)
RETURNING id, name, afl_player_id;

-- name: UpdatePlayer :one
UPDATE ffl.player
SET name = $2,
    afl_player_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, name, afl_player_id;

-- name: DeletePlayer :exec
UPDATE ffl.player
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
