-- Drop indexes first
DROP INDEX IF EXISTS ffl.idx_player_club_id;
DROP INDEX IF EXISTS ffl.idx_club_deleted_at;
DROP INDEX IF EXISTS ffl.idx_player_deleted_at;

-- Drop tables
DROP TABLE IF EXISTS ffl.player;
DROP TABLE IF EXISTS ffl.club;

-- Drop schema
DROP SCHEMA IF EXISTS ffl; 