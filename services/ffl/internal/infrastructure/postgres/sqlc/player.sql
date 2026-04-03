-- name: FindAllPlayers :many
SELECT id, drv_name, afl_player_id
FROM ffl.player
WHERE deleted_at IS NULL
ORDER BY drv_name;

-- name: FindPlayerByID :one
SELECT id, drv_name, afl_player_id
FROM ffl.player
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePlayer :one
INSERT INTO ffl.player (drv_name, afl_player_id)
VALUES ($1, $2)
RETURNING id, drv_name, afl_player_id;

-- name: UpdatePlayer :one
UPDATE ffl.player
SET drv_name = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, drv_name, afl_player_id;

-- name: FindPlayerByAFLPlayerID :one
SELECT id, drv_name, afl_player_id
FROM ffl.player
WHERE afl_player_id = $1 AND deleted_at IS NULL;

-- name: DeletePlayer :exec
UPDATE ffl.player
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
