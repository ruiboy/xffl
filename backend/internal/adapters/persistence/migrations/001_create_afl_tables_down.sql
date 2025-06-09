-- Drop indexes first
DROP INDEX IF EXISTS afl.idx_afl_season_league_id;
DROP INDEX IF EXISTS afl.idx_afl_round_season_id;
DROP INDEX IF EXISTS afl.idx_afl_match_round_id;
DROP INDEX IF EXISTS afl.idx_afl_match_home_club_match_id;
DROP INDEX IF EXISTS afl.idx_afl_match_away_club_match_id;
DROP INDEX IF EXISTS afl.idx_afl_club_season_club_id;
DROP INDEX IF EXISTS afl.idx_afl_club_season_season_id;
DROP INDEX IF EXISTS afl.idx_afl_club_match_match_id;
DROP INDEX IF EXISTS afl.idx_afl_club_match_club_season_id;
DROP INDEX IF EXISTS afl.idx_afl_player_club_id;
DROP INDEX IF EXISTS afl.idx_afl_player_season_player_id;
DROP INDEX IF EXISTS afl.idx_afl_player_season_club_season_id;
DROP INDEX IF EXISTS afl.idx_afl_player_match_club_match_id;
DROP INDEX IF EXISTS afl.idx_afl_player_match_player_season_id;
DROP INDEX IF EXISTS afl.idx_afl_league_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_season_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_round_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_match_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_club_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_club_season_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_club_match_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_player_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_player_season_deleted_at;
DROP INDEX IF EXISTS afl.idx_afl_player_match_deleted_at;

-- Drop foreign key constraints
ALTER TABLE afl.match DROP CONSTRAINT IF EXISTS fk_afl_match_home_club_match;
ALTER TABLE afl.match DROP CONSTRAINT IF EXISTS fk_afl_match_away_club_match;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS afl.player_match;
DROP TABLE IF EXISTS afl.player_season;
DROP TABLE IF EXISTS afl.player;
DROP TABLE IF EXISTS afl.club_match;
DROP TABLE IF EXISTS afl.club_season;
DROP TABLE IF EXISTS afl.club;
DROP TABLE IF EXISTS afl.match;
DROP TABLE IF EXISTS afl.round;
DROP TABLE IF EXISTS afl.season;
DROP TABLE IF EXISTS afl.league;

-- Drop schema
DROP SCHEMA IF EXISTS afl;