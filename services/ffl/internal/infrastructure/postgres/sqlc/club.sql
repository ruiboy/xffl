-- name: FindAllClubs :many
SELECT id, name
FROM ffl.club
WHERE deleted_at IS NULL
ORDER BY name;

-- name: FindClubByID :one
SELECT id, name
FROM ffl.club
WHERE id = $1 AND deleted_at IS NULL;
