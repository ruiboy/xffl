-- name: FindPlayerMatchesByClubMatchID :many
SELECT id, club_match_id, player_season_id,
       kicks, handballs, marks, hitouts, tackles, goals, behinds
FROM afl.player_match
WHERE club_match_id = $1 AND deleted_at IS NULL;

-- name: FindPlayerMatchByID :one
SELECT id, club_match_id, player_season_id,
       kicks, handballs, marks, hitouts, tackles, goals, behinds
FROM afl.player_match
WHERE id = $1 AND deleted_at IS NULL;
