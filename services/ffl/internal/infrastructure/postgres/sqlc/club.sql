-- name: FindAllClubs :many
SELECT id, name
FROM ffl.club
WHERE deleted_at IS NULL
ORDER BY name;

-- name: FindClubByID :one
SELECT id, name
FROM ffl.club
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindClubsByIDs :many
SELECT id, name
FROM ffl.club
WHERE id = ANY(@ids::int[]) AND deleted_at IS NULL;

-- name: FindClubByName :one
SELECT id, name FROM ffl.club WHERE name = $1 AND deleted_at IS NULL;

-- name: CreateClub :one
INSERT INTO ffl.club (name) VALUES ($1) RETURNING id, name;
