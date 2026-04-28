-- name: FindPlayerSeasonByID :one
SELECT id, player_id, club_season_id, from_round_id, to_round_id
FROM afl.player_season
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindPlayerSeasonsByClubSeasonIDWithPlayer :many
SELECT ps.id AS player_season_id, p.id AS player_id, p.name AS player_name, ps.club_season_id
FROM afl.player_season ps
JOIN afl.player p ON p.id = ps.player_id
WHERE ps.club_season_id = $1 AND ps.deleted_at IS NULL AND p.deleted_at IS NULL;

-- name: FindPlayersByPlayerSeasonIDs :many
SELECT ps.id AS player_season_id, p.id AS player_id, p.name AS player_name
FROM afl.player_season ps
JOIN afl.player p ON p.id = ps.player_id
WHERE ps.id = ANY(@player_season_ids::int[]) AND ps.deleted_at IS NULL;
