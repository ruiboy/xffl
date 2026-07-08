-- Queries for the one-time historical import (cmd/afltables-import).
-- Get-or-create helpers for the season→player_match scaffold, plus name lookups
-- used by the interactive player resolver.

-- name: UpsertLeagueByName :one
INSERT INTO afl.league (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: FindSeasonByLeagueAndName :one
SELECT id FROM afl.season
WHERE league_id = $1 AND name = $2 AND deleted_at IS NULL;

-- name: InsertSeasonReturningID :one
INSERT INTO afl.season (league_id, name) VALUES ($1, $2)
RETURNING id;

-- name: FindRoundBySeasonAndName :one
SELECT id FROM afl.round
WHERE season_id = $1 AND name = $2 AND deleted_at IS NULL;

-- name: InsertRoundReturningID :one
INSERT INTO afl.round (season_id, name) VALUES ($1, $2)
RETURNING id;

-- name: UpsertClubByName :one
INSERT INTO afl.club (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: UpsertClubSeasonReturningID :one
INSERT INTO afl.club_season (club_id, season_id) VALUES ($1, $2)
ON CONFLICT (club_id, season_id) DO UPDATE SET club_id = EXCLUDED.club_id
RETURNING id;

-- name: FindMatchByRoundAndHomeClubSeason :one
SELECT m.id FROM afl.match m
JOIN afl.club_match cm ON cm.match_id = m.id AND cm.side = 'home' AND cm.deleted_at IS NULL
WHERE m.round_id = $1 AND cm.club_season_id = $2 AND m.deleted_at IS NULL
LIMIT 1;

-- name: InsertHistoricalMatch :one
-- Historical matches are complete, so data_status is 'final' (so their stats
-- count in player-season averages). Scores/results are not backfilled.
INSERT INTO afl.match (round_id, venue, start_dt, data_status)
VALUES ($1, $2, $3, 'final')
RETURNING id;

-- name: UpsertClubMatchReturningID :one
INSERT INTO afl.club_match (match_id, club_season_id, side)
VALUES ($1, $2, $3)
ON CONFLICT (club_season_id, match_id) DO UPDATE SET side = EXCLUDED.side
RETURNING id;

-- name: FindPlayersByExactName :many
SELECT id, name FROM afl.player
WHERE name = $1 AND deleted_at IS NULL
ORDER BY id;

-- name: FindAllPlayers :many
SELECT id, name FROM afl.player
WHERE deleted_at IS NULL
ORDER BY id;

-- name: FindClubsForNamedPlayers :many
-- (player_id, club_name) pairs for every player sharing the given name — used
-- to disambiguate same-name players by the club of the row being imported.
SELECT p.id AS player_id, c.name AS club_name
FROM afl.player p
JOIN afl.player_season ps ON ps.player_id = p.id AND ps.deleted_at IS NULL
JOIN afl.club_season cs ON cs.id = ps.club_season_id AND cs.deleted_at IS NULL
JOIN afl.club c ON c.id = cs.club_id AND c.deleted_at IS NULL
WHERE p.name = $1 AND p.deleted_at IS NULL;

-- name: FindSeasonNamesByPlayerID :many
-- Distinct AFL season names a player has any player_season in — used to detect
-- career gaps when the same name recurs in non-consecutive seasons.
SELECT DISTINCT s.name
FROM afl.season s
JOIN afl.club_season cs ON cs.season_id = s.id AND cs.deleted_at IS NULL
JOIN afl.player_season ps ON ps.club_season_id = cs.id AND ps.deleted_at IS NULL
WHERE ps.player_id = $1 AND s.deleted_at IS NULL;
