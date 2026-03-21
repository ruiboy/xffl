-- name: FindClubMatchesByMatchID :many
SELECT id, match_id, club_season_id, rushed_behinds, drv_score
FROM afl.club_match
WHERE match_id = $1 AND deleted_at IS NULL;

-- name: FindClubMatchByID :one
SELECT id, match_id, club_season_id, rushed_behinds, drv_score
FROM afl.club_match
WHERE id = $1 AND deleted_at IS NULL;

-- name: RecalculateClubMatchScore :exec
UPDATE afl.club_match
SET drv_score = (
    SELECT COALESCE(SUM(COALESCE(goals, 0) * 6 + COALESCE(behinds, 0)), 0)
    FROM afl.player_match
    WHERE club_match_id = $1 AND deleted_at IS NULL
) + COALESCE(rushed_behinds, 0),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
