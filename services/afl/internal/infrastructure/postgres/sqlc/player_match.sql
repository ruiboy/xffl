-- name: FindPlayerMatchesByClubMatchID :many
SELECT pm.id, pm.club_match_id, pm.player_season_id,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals, pm.behinds,
       m.data_status
FROM afl.player_match pm
JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
WHERE pm.club_match_id = $1 AND pm.deleted_at IS NULL;

-- name: FindPlayerMatchByID :one
SELECT pm.id, pm.club_match_id, pm.player_season_id,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals, pm.behinds,
       m.data_status
FROM afl.player_match pm
JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
WHERE pm.id = $1 AND pm.deleted_at IS NULL;

-- name: FindPlayerMatchesByPlayerSeasonID :many
SELECT pm.id, pm.club_match_id, pm.player_season_id,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals, pm.behinds,
       m.data_status
FROM afl.player_match pm
JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
WHERE pm.player_season_id = $1 AND pm.deleted_at IS NULL
ORDER BY pm.id;

-- name: FindPlayerMatchesByIDs :many
SELECT pm.id, pm.club_match_id, pm.player_season_id,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals, pm.behinds,
       m.data_status
FROM afl.player_match pm
JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
WHERE pm.id = ANY(@ids::int[]) AND pm.deleted_at IS NULL;

-- name: FindPlayerMatchesBySeasonIDsAndRoundID :many
SELECT pm.id, pm.club_match_id, pm.player_season_id,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals, pm.behinds,
       m.data_status
FROM afl.player_match pm
JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
WHERE pm.player_season_id = ANY(@player_season_ids::int[])
  AND m.round_id = @round_id
  AND pm.deleted_at IS NULL;

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
