-- name: FindClubMatchesByMatchID :many
SELECT id, match_id, club_season_id, data_status, notes, drv_score
FROM ffl.club_match
WHERE match_id = $1 AND deleted_at IS NULL
ORDER BY CASE WHEN side = 'home' THEN 0 ELSE 1 END;

-- name: FindClubMatchByID :one
SELECT id, match_id, club_season_id, data_status, notes, drv_score
FROM ffl.club_match
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindClubMatchesByIDs :many
SELECT id, match_id, club_season_id, data_status, notes, drv_score
FROM ffl.club_match
WHERE id = ANY($1::int[]) AND deleted_at IS NULL;

-- name: UpdateClubMatchScore :exec
UPDATE ffl.club_match
SET drv_score = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateClubMatchNotes :exec
UPDATE ffl.club_match
SET notes      = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateClubMatchDataStatus :exec
UPDATE ffl.club_match
SET data_status = $2,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateClubMatchPremiershipPoints :exec
UPDATE ffl.club_match
SET drv_premiership_points = $2,
    updated_at             = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CountFinalClubMatchesByMatchID :one
SELECT COUNT(*) FROM ffl.club_match
WHERE match_id = $1 AND data_status = 'final' AND deleted_at IS NULL;

-- name: GetRulesIDByClubMatchID :one
SELECT s.rules_id
FROM ffl.club_match cm
JOIN ffl.match m  ON m.id = cm.match_id
JOIN ffl.round r  ON r.id = m.round_id
JOIN ffl.season s ON s.id = r.season_id
WHERE cm.id = $1 AND cm.deleted_at IS NULL;

-- name: CreateClubMatch :one
INSERT INTO ffl.club_match (match_id, club_season_id, side)
VALUES ($1, $2, $3)
RETURNING id, match_id, club_season_id, side, data_status;

-- name: SoftDeleteClubMatchesByMatchID :exec
UPDATE ffl.club_match
SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE match_id = $1 AND deleted_at IS NULL;

-- name: FindFinalFflByesBySeasonID :many
SELECT cm.club_season_id, COALESCE(cm.drv_score, 0) AS score, r.round_type
FROM ffl.club_match cm
JOIN ffl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
JOIN ffl.round r ON r.id = m.round_id AND r.deleted_at IS NULL
WHERE r.season_id = $1 AND cm.side = 'bye' AND cm.data_status = 'final' AND cm.deleted_at IS NULL;
