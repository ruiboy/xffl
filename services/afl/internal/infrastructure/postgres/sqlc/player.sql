-- name: FindPlayerByID :one
SELECT id, name
FROM afl.player
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindPlayersByIDs :many
SELECT id, name
FROM afl.player
WHERE id = ANY(@ids::int[]) AND deleted_at IS NULL
ORDER BY id;

-- name: SearchPlayersByName :many
SELECT id, name
FROM afl.player
WHERE name ILIKE '%' || @query || '%' AND deleted_at IS NULL
ORDER BY name
LIMIT 20;
