-- name: FindAllPlayers :many
SELECT id, name
FROM ffl.player
WHERE deleted_at IS NULL
ORDER BY name;

-- name: FindPlayerByID :one
SELECT id, name
FROM ffl.player
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreatePlayer :one
INSERT INTO ffl.player (name)
VALUES ($1)
RETURNING id, name;

-- name: UpdatePlayer :one
UPDATE ffl.player
SET name = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, name;

-- name: DeletePlayer :exec
UPDATE ffl.player
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
