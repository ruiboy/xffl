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
