-- name: FindRoundsBySeasonID :many
SELECT r.id, r.name, r.season_id
FROM afl.round r
LEFT JOIN afl.match m ON m.round_id = r.id AND m.deleted_at IS NULL
WHERE r.season_id = $1 AND r.deleted_at IS NULL
GROUP BY r.id, r.name, r.season_id
ORDER BY MIN(m.start_dt) NULLS LAST, r.id;

-- name: FindRoundByID :one
SELECT id, name, season_id
FROM afl.round
WHERE id = $1 AND deleted_at IS NULL;

