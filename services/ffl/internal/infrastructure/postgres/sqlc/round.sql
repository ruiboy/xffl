-- name: FindRoundsBySeasonID :many
SELECT id, name, season_id
FROM ffl.round
WHERE season_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: FindRoundByID :one
SELECT id, name, season_id
FROM ffl.round
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindLatestRound :one
SELECT r.id, r.name, r.season_id
FROM ffl.round r
JOIN ffl.season s ON s.id = r.season_id AND s.deleted_at IS NULL
WHERE r.deleted_at IS NULL
ORDER BY s.id DESC, r.id DESC
LIMIT 1;
