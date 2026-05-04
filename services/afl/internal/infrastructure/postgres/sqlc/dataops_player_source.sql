-- name: FindDataopsPlayerSource :one
SELECT player_season_id
FROM afl.dataops_player_source
WHERE source = $1
  AND external_season = $2
  AND external_club = $3
  AND external_player = $4;

-- name: UpsertDataopsPlayerSource :exec
INSERT INTO afl.dataops_player_source (source, external_season, external_club, external_player, player_season_id, updated_at)
VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
ON CONFLICT (source, external_season, external_club, external_player) DO UPDATE
    SET player_season_id = EXCLUDED.player_season_id,
        updated_at       = CURRENT_TIMESTAMP;
