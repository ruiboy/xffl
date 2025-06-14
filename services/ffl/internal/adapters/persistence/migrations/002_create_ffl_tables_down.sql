-- Drop indexes first
DROP INDEX IF EXISTS ffl.idx_season_league_id;
DROP INDEX IF EXISTS ffl.idx_round_season_id;
DROP INDEX IF EXISTS ffl.idx_match_round_id;
DROP INDEX IF EXISTS ffl.idx_match_home_club_match_id;
DROP INDEX IF EXISTS ffl.idx_match_away_club_match_id;
DROP INDEX IF EXISTS ffl.idx_club_season_club_id;
DROP INDEX IF EXISTS ffl.idx_club_season_season_id;
DROP INDEX IF EXISTS ffl.idx_club_match_match_id;
DROP INDEX IF EXISTS ffl.idx_club_match_club_season_id;
DROP INDEX IF EXISTS ffl.idx_player_club_id;
DROP INDEX IF EXISTS ffl.idx_player_season_player_id;
DROP INDEX IF EXISTS ffl.idx_player_season_club_season_id;
DROP INDEX IF EXISTS ffl.idx_player_season_from_round_id;
DROP INDEX IF EXISTS ffl.idx_player_season_to_round_id;
DROP INDEX IF EXISTS ffl.idx_player_match_club_match_id;
DROP INDEX IF EXISTS ffl.idx_player_match_player_season_id;
DROP INDEX IF EXISTS ffl.idx_league_deleted_at;
DROP INDEX IF EXISTS ffl.idx_season_deleted_at;
DROP INDEX IF EXISTS ffl.idx_round_deleted_at;
DROP INDEX IF EXISTS ffl.idx_match_deleted_at;
DROP INDEX IF EXISTS ffl.idx_club_deleted_at;
DROP INDEX IF EXISTS ffl.idx_club_season_deleted_at;
DROP INDEX IF EXISTS ffl.idx_club_match_deleted_at;
DROP INDEX IF EXISTS ffl.idx_player_deleted_at;
DROP INDEX IF EXISTS ffl.idx_player_season_deleted_at;
DROP INDEX IF EXISTS ffl.idx_player_match_deleted_at;

-- Drop foreign key constraints
ALTER TABLE ffl.match DROP CONSTRAINT IF EXISTS fk_match_home_club_match;
ALTER TABLE ffl.match DROP CONSTRAINT IF EXISTS fk_match_away_club_match;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS ffl.player_match;
DROP TABLE IF EXISTS ffl.player_season;
DROP TABLE IF EXISTS ffl.player;
DROP TABLE IF EXISTS ffl.club_match;
DROP TABLE IF EXISTS ffl.club_season;
DROP TABLE IF EXISTS ffl.club;
DROP TABLE IF EXISTS ffl.match;
DROP TABLE IF EXISTS ffl.round;
DROP TABLE IF EXISTS ffl.season;
DROP TABLE IF EXISTS ffl.league;

-- Drop schema
DROP SCHEMA IF EXISTS ffl;