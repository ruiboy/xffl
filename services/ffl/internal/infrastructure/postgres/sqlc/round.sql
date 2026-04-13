-- name: FindRoundsBySeasonID :many
SELECT r.id, r.name, r.season_id
FROM ffl.round r
LEFT JOIN ffl.match m ON m.round_id = r.id AND m.deleted_at IS NULL
WHERE r.season_id = $1 AND r.deleted_at IS NULL
GROUP BY r.id, r.name, r.season_id
ORDER BY MIN(m.start_dt) NULLS LAST, r.id;

-- name: FindRoundByID :one
SELECT id, name, season_id
FROM ffl.round
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindRoundByAFLRoundID :one
SELECT id, name, season_id, afl_round_id
FROM ffl.round
WHERE afl_round_id = $1 AND deleted_at IS NULL;
