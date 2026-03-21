-- name: FindPlayersByClubID :many
SELECT DISTINCT p.id, p.name, COALESCE(p.club_id, 0) AS club_id
FROM afl.player p
LEFT JOIN afl.player_season ps ON ps.player_id = p.id AND ps.deleted_at IS NULL
LEFT JOIN afl.club_season cs ON ps.club_season_id = cs.id AND cs.deleted_at IS NULL
WHERE (p.club_id = $1 OR cs.club_id = $1) AND p.deleted_at IS NULL
ORDER BY p.name;

-- name: FindPlayerByID :one
SELECT id, name, COALESCE(club_id, 0) AS club_id
FROM afl.player
WHERE id = $1 AND deleted_at IS NULL;
