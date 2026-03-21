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

-- name: UpsertPlayerMatch :one
INSERT INTO afl.player_match (club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (player_season_id, club_match_id)
DO UPDATE SET
    kicks = COALESCE($3, afl.player_match.kicks),
    handballs = COALESCE($4, afl.player_match.handballs),
    marks = COALESCE($5, afl.player_match.marks),
    hitouts = COALESCE($6, afl.player_match.hitouts),
    tackles = COALESCE($7, afl.player_match.tackles),
    goals = COALESCE($8, afl.player_match.goals),
    behinds = COALESCE($9, afl.player_match.behinds),
    updated_at = CURRENT_TIMESTAMP
WHERE afl.player_match.deleted_at IS NULL
RETURNING id, club_match_id, player_season_id, kicks, handballs, marks, hitouts, tackles, goals, behinds;
