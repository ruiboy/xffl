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

-- name: FindLatestRound :one
SELECT r.id, r.name, r.season_id
FROM afl.round r
JOIN afl.season s ON s.id = r.season_id AND s.deleted_at IS NULL
WHERE r.deleted_at IS NULL
ORDER BY s.id DESC, r.id DESC
LIMIT 1;

-- name: FindRoundsWithMatchBoundsBySeasonID :many
SELECT
    r.id,
    r.name,
    r.season_id,
    COALESCE(MIN(m.start_dt), '0001-01-01T00:00:00Z'::timestamptz) AS first_match_dt,
    COALESCE(MAX(m.start_dt), '0001-01-01T00:00:00Z'::timestamptz) AS last_match_dt
FROM afl.round r
JOIN afl.match m ON m.round_id = r.id AND m.deleted_at IS NULL
WHERE r.season_id = $1 AND r.deleted_at IS NULL
GROUP BY r.id, r.name, r.season_id
ORDER BY MIN(m.start_dt);
