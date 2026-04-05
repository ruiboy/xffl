-- name: FindPlayerMatchesByClubMatchID :many
SELECT id, club_match_id, player_season_id,
       position, status, backup_positions, interchange_position, drv_score, afl_player_match_id
FROM ffl.player_match
WHERE club_match_id = $1 AND deleted_at IS NULL;

-- name: FindPlayerMatchByID :one
SELECT id, club_match_id, player_season_id,
       position, status, backup_positions, interchange_position, drv_score, afl_player_match_id
FROM ffl.player_match
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeletePlayerMatchesByClubMatchID :exec
DELETE FROM ffl.player_match
WHERE club_match_id = $1;

-- name: UpsertPlayerMatch :one
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, backup_positions, interchange_position, drv_score)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (player_season_id, club_match_id)
DO UPDATE SET
    position = COALESCE($3, ffl.player_match.position),
    status = COALESCE($4, ffl.player_match.status),
    backup_positions = $5,
    interchange_position = $6,
    drv_score = COALESCE($7, ffl.player_match.drv_score),
    updated_at = CURRENT_TIMESTAMP
WHERE ffl.player_match.deleted_at IS NULL
RETURNING id, club_match_id, player_season_id, position, status, backup_positions, interchange_position, drv_score, afl_player_match_id;
