-- name: FindPlayerMatchesByClubMatchID :many
SELECT id, club_match_id, player_season_id, status,
       kicks, handballs, marks, hitouts, tackles, goals, behinds
FROM afl.player_match
WHERE club_match_id = $1 AND deleted_at IS NULL;

-- name: FindPlayerMatchByID :one
SELECT id, club_match_id, player_season_id, status,
       kicks, handballs, marks, hitouts, tackles, goals, behinds
FROM afl.player_match
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindPlayerMatchesByPlayerSeasonID :many
SELECT id, club_match_id, player_season_id, status,
       kicks, handballs, marks, hitouts, tackles, goals, behinds
FROM afl.player_match
WHERE player_season_id = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: FindPlayerMatchesByIDs :many
SELECT id, club_match_id, player_season_id, status,
       kicks, handballs, marks, hitouts, tackles, goals, behinds
FROM afl.player_match
WHERE id = ANY(@ids::int[]) AND deleted_at IS NULL;

-- name: FindPlayerMatchesBySeasonIDsAndRoundID :many
SELECT pm.id, pm.club_match_id, pm.player_season_id, pm.status,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals, pm.behinds
FROM afl.player_match pm
JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
WHERE pm.player_season_id = ANY(@player_season_ids::int[])
  AND m.round_id = @round_id
  AND pm.deleted_at IS NULL;

-- name: SetPlayerMatchStatusForMatch :exec
UPDATE afl.player_match
SET status = $2, updated_at = CURRENT_TIMESTAMP
FROM afl.club_match cm
WHERE afl.player_match.club_match_id = cm.id
  AND cm.match_id = $1
  AND afl.player_match.deleted_at IS NULL;

-- name: UpsertPlayerMatch :one
INSERT INTO afl.player_match (club_match_id, player_season_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (player_season_id, club_match_id)
DO UPDATE SET
    status = COALESCE($3, afl.player_match.status),
    kicks = COALESCE($4, afl.player_match.kicks),
    handballs = COALESCE($5, afl.player_match.handballs),
    marks = COALESCE($6, afl.player_match.marks),
    hitouts = COALESCE($7, afl.player_match.hitouts),
    tackles = COALESCE($8, afl.player_match.tackles),
    goals = COALESCE($9, afl.player_match.goals),
    behinds = COALESCE($10, afl.player_match.behinds),
    updated_at = CURRENT_TIMESTAMP
WHERE afl.player_match.deleted_at IS NULL
RETURNING id, club_match_id, player_season_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds;
