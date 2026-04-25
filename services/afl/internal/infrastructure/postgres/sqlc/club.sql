-- name: FindAllClubs :many
SELECT id, name
FROM afl.club
WHERE deleted_at IS NULL
ORDER BY name;

-- name: FindClubByID :one
SELECT id, name
FROM afl.club
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindClubsByIDs :many
SELECT id, name
FROM afl.club
WHERE id = ANY(@ids::int[]) AND deleted_at IS NULL;
