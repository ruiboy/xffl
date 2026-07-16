-- name: FindAllSeasons :many
SELECT id, name, league_id, afl_season_id, rules_id
FROM ffl.season
WHERE deleted_at IS NULL
ORDER BY name;

-- name: FindSeasonByID :one
SELECT id, name, league_id, afl_season_id, rules_id
FROM ffl.season
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateSeason :one
INSERT INTO ffl.season (league_id, name, afl_season_id, rules_id)
VALUES ($1, $2, $3, $4)
RETURNING id, name, league_id, afl_season_id, rules_id;
