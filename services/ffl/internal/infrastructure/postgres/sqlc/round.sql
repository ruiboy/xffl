-- Rounds run in season order. start_dt is the real key, but the fixture builder
-- never sets it (an FFL round's timing comes from the AFL round it maps to, and
-- that lives in the afl schema, which this one does not join to), so every
-- builder-made round ties on NULL. afl_round_id breaks that tie: a season's AFL
-- rounds are created in sequence, so ascending id is season order. Falling back
-- to r.id alone would sort by creation, putting a round added late last.
-- name: FindRoundsBySeasonID :many
SELECT r.id, r.name, r.season_id, r.afl_round_id, r.round_type
FROM ffl.round r
LEFT JOIN ffl.match m ON m.round_id = r.id AND m.deleted_at IS NULL
WHERE r.season_id = $1 AND r.deleted_at IS NULL
GROUP BY r.id, r.name, r.season_id, r.afl_round_id, r.round_type
ORDER BY MIN(m.start_dt) NULLS LAST, r.afl_round_id, r.id;

-- name: FindRoundByID :one
SELECT id, name, season_id, afl_round_id, round_type
FROM ffl.round
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindRoundByAFLRoundID :one
SELECT id, name, season_id, afl_round_id
FROM ffl.round
WHERE afl_round_id = $1 AND deleted_at IS NULL;

-- name: CreateRound :one
INSERT INTO ffl.round (season_id, name, afl_round_id, round_type)
VALUES ($1, $2, $3, $4)
RETURNING id, name, season_id, afl_round_id, round_type;

-- name: UpdateRound :exec
UPDATE ffl.round
SET name = $2, afl_round_id = $3, round_type = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteRound :exec
UPDATE ffl.round
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
