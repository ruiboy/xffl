-- name: FindClubMatchesByMatchID :many
SELECT id, match_id, club_season_id, drv_score
FROM ffl.club_match
WHERE match_id = $1 AND deleted_at IS NULL;

-- name: FindClubMatchByID :one
SELECT id, match_id, club_season_id, drv_score
FROM ffl.club_match
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateClubMatchScore :exec
UPDATE ffl.club_match
SET drv_score = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
