-- name: FindPlayerSeasonByID :one
SELECT id, player_id, club_season_id, from_round_id, to_round_id
FROM afl.player_season
WHERE id = $1 AND deleted_at IS NULL;
