-- name: FindClubMatchesByMatchID :many
SELECT id, match_id, club_season_id, rushed_behinds, drv_score
FROM afl.club_match
WHERE match_id = $1 AND deleted_at IS NULL;

-- name: FindClubMatchByID :one
SELECT id, match_id, club_season_id, rushed_behinds, drv_score
FROM afl.club_match
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateClubMatchScore :exec
UPDATE afl.club_match
SET drv_score = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindRoundIDByClubMatchID :one
SELECT m.round_id
FROM afl.match m
JOIN afl.club_match cm ON cm.match_id = m.id
WHERE cm.id = $1 AND cm.deleted_at IS NULL AND m.deleted_at IS NULL;
