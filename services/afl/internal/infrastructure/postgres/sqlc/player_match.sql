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

-- name: FindPlayerMatchStatsByPlayerSeasonIDs :many
SELECT player_season_id,
       COUNT(*)::INTEGER AS games_played,
       AVG(kicks)::FLOAT8 AS avg_kicks,
       AVG(handballs)::FLOAT8 AS avg_handballs,
       AVG(marks)::FLOAT8 AS avg_marks,
       AVG(hitouts)::FLOAT8 AS avg_hitouts,
       AVG(tackles)::FLOAT8 AS avg_tackles,
       AVG(goals)::FLOAT8 AS avg_goals,
       AVG(behinds)::FLOAT8 AS avg_behinds
FROM afl.player_match
WHERE player_season_id = ANY(@player_season_ids::int[]) AND deleted_at IS NULL
GROUP BY player_season_id;

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
