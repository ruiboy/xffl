-- name: FindAllPlayers :many
SELECT id, afl_player_id
FROM ffl.player
WHERE deleted_at IS NULL
ORDER BY id;

-- name: FindPlayerByID :one
SELECT id, afl_player_id
FROM ffl.player
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePlayer :one
INSERT INTO ffl.player (afl_player_id)
VALUES ($1)
RETURNING id, afl_player_id;

-- name: FindPlayerByAFLPlayerID :one
SELECT id, afl_player_id
FROM ffl.player
WHERE afl_player_id = $1 AND deleted_at IS NULL;

-- name: DeletePlayer :exec
UPDATE ffl.player
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
