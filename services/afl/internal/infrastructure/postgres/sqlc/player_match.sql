-- name: FindPlayerMatchesByClubMatchID :many
SELECT pm.id, pm.club_match_id, pm.player_season_id,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals, pm.behinds,
       m.data_status
FROM afl.player_match pm
JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
WHERE pm.club_match_id = $1 AND pm.deleted_at IS NULL
ORDER BY pm.id;

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
WHERE pm.id = ANY(@ids::int[]) AND pm.deleted_at IS NULL
ORDER BY pm.id;

-- name: FindPlayerMatchesBySeasonIDsAndRoundID :many
SELECT pm.id, pm.club_match_id, pm.player_season_id,
       pm.kicks, pm.handballs, pm.marks, pm.hitouts, pm.tackles, pm.goals, pm.behinds,
       m.data_status
FROM afl.player_match pm
JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
WHERE pm.player_season_id = ANY(@player_season_ids::int[])
  AND m.round_id = @round_id
  AND pm.deleted_at IS NULL
ORDER BY pm.id;

-- name: FindByeStatusBatch :many
-- For each player_season_id, returns whether their club has a bye in the given
-- round, and whether they played in their club's most recent non-bye final match
-- before that round.
SELECT
  ps.id AS player_season_id,
  EXISTS (
    SELECT 1 FROM afl.bye b
    WHERE b.round_id = @round_id
      AND b.club_season_id = ps.club_season_id
      AND b.deleted_at IS NULL
  ) AS has_bye,
  EXISTS (
    SELECT 1
    FROM afl.player_match pm
    JOIN afl.club_match cm ON cm.id = pm.club_match_id AND cm.deleted_at IS NULL
    JOIN afl.match m ON m.id = cm.match_id AND m.deleted_at IS NULL
    WHERE pm.player_season_id = ps.id
      AND pm.deleted_at IS NULL
      AND m.round_id = (
        SELECT r.id
        FROM afl.match m2
        JOIN afl.club_match cm2 ON cm2.match_id = m2.id AND cm2.deleted_at IS NULL
        JOIN afl.round r ON r.id = m2.round_id AND r.deleted_at IS NULL
        WHERE cm2.club_season_id = ps.club_season_id
          AND r.id < @round_id
          AND m2.deleted_at IS NULL
          AND m2.data_status = 'final'
          AND NOT EXISTS (
            SELECT 1 FROM afl.bye b2
            WHERE b2.round_id = r.id
              AND b2.club_season_id = ps.club_season_id
              AND b2.deleted_at IS NULL
          )
        ORDER BY r.id DESC
        LIMIT 1
      )
  ) AS played_last
FROM afl.player_season ps
WHERE ps.id = ANY(@player_season_ids::int[])
  AND ps.deleted_at IS NULL;

-- name: GetPlayerSeasonAveragesBatch :many
-- Returns season-to-date average stats for each player_season, across all matches played.
SELECT
  pm.player_season_id,
  AVG(pm.goals)::float8     AS avg_goals,
  AVG(pm.kicks)::float8     AS avg_kicks,
  AVG(pm.handballs)::float8 AS avg_handballs,
  AVG(pm.marks)::float8     AS avg_marks,
  AVG(pm.tackles)::float8   AS avg_tackles,
  AVG(pm.hitouts)::float8   AS avg_hitouts
FROM afl.player_match pm
WHERE pm.player_season_id = ANY(@player_season_ids::int[])
  AND pm.deleted_at IS NULL
GROUP BY pm.player_season_id;

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
