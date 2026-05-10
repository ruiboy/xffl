-- name: FindClubSeasonsBySeasonID :many
SELECT id, club_id, season_id,
       drv_played, drv_won, drv_lost, drv_drawn,
       drv_for, drv_against, drv_premiership_points
FROM afl.club_season
WHERE season_id = $1 AND deleted_at IS NULL
ORDER BY drv_premiership_points DESC, (drv_for - drv_against) DESC;

-- name: FindClubSeasonByID :one
SELECT id, club_id, season_id,
       drv_played, drv_won, drv_lost, drv_drawn,
       drv_for, drv_against, drv_premiership_points
FROM afl.club_season
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateClubSeason :exec
UPDATE afl.club_season
SET drv_played             = $2,
    drv_won                = $3,
    drv_lost               = $4,
    drv_drawn              = $5,
    drv_for                = $6,
    drv_against            = $7,
    drv_premiership_points = $8,
    updated_at             = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;
