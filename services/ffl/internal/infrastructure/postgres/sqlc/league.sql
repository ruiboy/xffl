-- name: FindAllLeagues :many
SELECT id, name FROM ffl.league WHERE deleted_at IS NULL ORDER BY id;

-- name: CreateLeague :one
INSERT INTO ffl.league (name) VALUES ($1) RETURNING id, name;
