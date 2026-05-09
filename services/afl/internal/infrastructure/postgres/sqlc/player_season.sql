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

-- name: FindPlayerSeasonsByIDs :many
SELECT id, player_id, club_season_id, from_round_id, to_round_id
FROM afl.player_season
WHERE id = ANY(@ids::int[]) AND deleted_at IS NULL;

-- name: InsertPlayerSeason :one
INSERT INTO afl.player_season (player_id, club_season_id)
VALUES ($1, $2)
RETURNING id, player_id, club_season_id, from_round_id, to_round_id;

-- name: UpsertPlayerSeason :one
INSERT INTO afl.player_season (player_id, club_season_id)
VALUES ($1, $2)
ON CONFLICT (player_id, club_season_id) DO UPDATE SET player_id = EXCLUDED.player_id
RETURNING id, player_id, club_season_id, from_round_id, to_round_id;

-- name: FindLatestPlayerSeasonByPlayerID :one
SELECT ps.id
FROM afl.player_season ps
JOIN afl.club_season cs ON cs.id = ps.club_season_id
WHERE ps.player_id = $1
  AND ps.deleted_at IS NULL
  AND cs.deleted_at IS NULL
ORDER BY cs.season_id DESC
LIMIT 1;

-- name: FindPlayerSeasonsBySeasonID :many
SELECT ps.id
FROM afl.player_season ps
JOIN afl.club_season cs ON cs.id = ps.club_season_id
JOIN afl.player p ON p.id = ps.player_id
WHERE cs.season_id = @season_id
  AND ps.deleted_at IS NULL
  AND cs.deleted_at IS NULL
  AND p.deleted_at IS NULL
  AND (sqlc.narg('name_query')::text IS NULL OR p.name ILIKE '%' || sqlc.narg('name_query') || '%')
ORDER BY p.name ASC, ps.id ASC;
