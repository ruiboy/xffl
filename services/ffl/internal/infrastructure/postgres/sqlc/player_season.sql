-- name: FindPlayersByPlayerSeasonIDs :many
SELECT ps.id AS player_season_id, p.id AS player_id, p.drv_name AS player_name
FROM ffl.player_season ps
JOIN ffl.player p ON p.id = ps.player_id
WHERE ps.id = ANY(@player_season_ids::int[]) AND ps.deleted_at IS NULL;

-- name: FindPlayerSeasonsByClubSeasonID :many
SELECT id, player_id, club_season_id, afl_player_season_id
FROM ffl.player_season
WHERE club_season_id = $1 AND deleted_at IS NULL;

-- name: FindPlayerSeasonByID :one
SELECT id, player_id, club_season_id, afl_player_season_id
FROM ffl.player_season
WHERE id = $1 AND deleted_at IS NULL;

-- name: FindPlayerSeasonsByAFLPlayerSeasonID :many
SELECT id, player_id, club_season_id, afl_player_season_id
FROM ffl.player_season
WHERE afl_player_season_id = $1 AND deleted_at IS NULL;

-- name: CreatePlayerSeason :one
INSERT INTO ffl.player_season (player_id, club_season_id)
VALUES ($1, $2)
RETURNING id, player_id, club_season_id, afl_player_season_id;

-- name: DeletePlayerSeason :exec
UPDATE ffl.player_season
SET deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
