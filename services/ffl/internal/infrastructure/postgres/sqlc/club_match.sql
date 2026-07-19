-- name: FindClubMatchesByMatchID :many
SELECT id, match_id, club_season_id, side, data_status, notes, drv_score
FROM ffl.club_match
WHERE match_id = $1 AND deleted_at IS NULL
ORDER BY CASE WHEN side = 'home' THEN 0 ELSE 1 END;

-- name: FindClubMatchByID :one
SELECT id, match_id, club_season_id, side, data_status, notes, drv_score
FROM ffl.club_match
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindClubMatchesByIDs :many
SELECT id, match_id, club_season_id, side, data_status, notes, drv_score
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

-- name: FindFinalFflClubMatchesBySeasonID :many
SELECT cm.match_id, cm.id AS club_match_id, cm.club_season_id,
       COALESCE(cm.drv_score, 0) AS score,
       COALESCE(m.match_style, '') AS match_style,
       r.round_type
FROM ffl.club_match cm
JOIN ffl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
JOIN ffl.round r ON r.id = m.round_id AND r.deleted_at IS NULL
WHERE r.season_id = $1 AND cm.data_status = 'final' AND cm.deleted_at IS NULL
ORDER BY cm.match_id;

-- name: UpdateClubMatchSide :exec
UPDATE ffl.club_match
SET side = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- Drops a club from a match outright. The fixture builder only edits rounds with
-- no submitted teams, so there is nothing here worth keeping — and a soft delete
-- would leave a tombstone that uni_ffl_club_match (which ignores deleted_at)
-- later blocks the same club from rejoining the match against.
-- name: DeleteClubMatchByID :exec
DELETE FROM ffl.club_match WHERE id = $1;
