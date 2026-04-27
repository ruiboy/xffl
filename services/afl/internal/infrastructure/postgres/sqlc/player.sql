-- name: FindPlayerByID :one
SELECT id, name
FROM afl.player
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindPlayersByIDs :many
SELECT id, name
FROM afl.player
WHERE id = ANY(@ids::int[]) AND deleted_at IS NULL
ORDER BY id;

-- name: FindPlayersByIDsWithClub :many
SELECT DISTINCT ON (p.id)
    p.id,
    p.name,
    COALESCE(c.name, '') AS club_name
FROM afl.player p
LEFT JOIN afl.player_season ps ON ps.player_id = p.id AND ps.deleted_at IS NULL
LEFT JOIN afl.club_season cs ON cs.id = ps.club_season_id AND cs.deleted_at IS NULL
LEFT JOIN afl.club c ON c.id = cs.club_id AND c.deleted_at IS NULL
WHERE p.id = ANY(@ids::int[]) AND p.deleted_at IS NULL
ORDER BY p.id, ps.id DESC;

-- name: SearchPlayersByName :many
SELECT id, name
FROM afl.player
WHERE name ILIKE '%' || @query || '%' AND deleted_at IS NULL
ORDER BY name
LIMIT 20;
