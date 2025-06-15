-- Create the afl schema
CREATE SCHEMA IF NOT EXISTS afl;

-- Create league table
CREATE TABLE IF NOT EXISTS afl.league (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT uni_afl_league_name UNIQUE (name)
);

-- Create season table
CREATE TABLE IF NOT EXISTS afl.season (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    league_id INTEGER NOT NULL REFERENCES afl.league(id) ON DELETE CASCADE
);

-- Create round table
CREATE TABLE IF NOT EXISTS afl.round (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    season_id INTEGER NOT NULL REFERENCES afl.season(id) ON DELETE CASCADE
);

-- Create match table
CREATE TABLE IF NOT EXISTS afl.match (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    round_id INTEGER NOT NULL REFERENCES afl.round(id) ON DELETE CASCADE,
    home_club_match_id INTEGER,
    away_club_match_id INTEGER,
    venue VARCHAR(255),
    start_dt TIMESTAMP WITH TIME ZONE,
    drv_result VARCHAR(50) CHECK (drv_result IN ('home_win', 'away_win', 'draw', 'no_result'))
);

-- Create club table
CREATE TABLE IF NOT EXISTS afl.club (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT uni_afl_club_name UNIQUE (name)
);

-- Create club_season table
CREATE TABLE IF NOT EXISTS afl.club_season (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    club_id INTEGER NOT NULL REFERENCES afl.club(id) ON DELETE CASCADE,
    season_id INTEGER NOT NULL REFERENCES afl.season(id) ON DELETE CASCADE,
    drv_played INTEGER DEFAULT 0,
    drv_won INTEGER DEFAULT 0,
    drv_lost INTEGER DEFAULT 0,
    drv_drawn INTEGER DEFAULT 0,
    drv_for INTEGER DEFAULT 0,
    drv_against INTEGER DEFAULT 0,
    drv_premiership_points INTEGER DEFAULT 0,
    CONSTRAINT uni_afl_club_season UNIQUE (club_id, season_id)
);

-- Create club_match table
CREATE TABLE IF NOT EXISTS afl.club_match (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    match_id INTEGER NOT NULL REFERENCES afl.match(id) ON DELETE CASCADE,
    club_season_id INTEGER NOT NULL REFERENCES afl.club_season(id) ON DELETE CASCADE,
    rushed_behinds INTEGER DEFAULT 0,
    drv_score INTEGER DEFAULT 0,
    drv_premiership_points INTEGER DEFAULT 0,
    CONSTRAINT uni_afl_club_match UNIQUE (club_season_id, match_id)
);


-- Add foreign key constraints for match table references to club_match
ALTER TABLE afl.match 
ADD CONSTRAINT fk_afl_match_home_club_match 
FOREIGN KEY (home_club_match_id) REFERENCES afl.club_match(id);

ALTER TABLE afl.match 
ADD CONSTRAINT fk_afl_match_away_club_match 
FOREIGN KEY (away_club_match_id) REFERENCES afl.club_match(id);

-- Create player table
CREATE TABLE IF NOT EXISTS afl.player (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    club_id INTEGER REFERENCES afl.club(id) ON DELETE CASCADE
);

-- Create player_season table
CREATE TABLE IF NOT EXISTS afl.player_season (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    player_id INTEGER NOT NULL REFERENCES afl.player(id) ON DELETE CASCADE,
    club_season_id INTEGER NOT NULL REFERENCES afl.club_season(id) ON DELETE CASCADE,
    CONSTRAINT uni_afl_player_season UNIQUE (player_id, club_season_id)
);

-- Create player_match table
CREATE TABLE IF NOT EXISTS afl.player_match (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    club_match_id INTEGER NOT NULL REFERENCES afl.club_match(id) ON DELETE CASCADE,
    player_season_id INTEGER NOT NULL REFERENCES afl.player_season(id) ON DELETE CASCADE,
    kicks INTEGER DEFAULT 0,
    handballs INTEGER DEFAULT 0,
    marks INTEGER DEFAULT 0,
    hitouts INTEGER DEFAULT 0,
    tackles INTEGER DEFAULT 0,
    goals INTEGER DEFAULT 0,
    behinds INTEGER DEFAULT 0,
    CONSTRAINT uni_afl_player_match UNIQUE (player_season_id, club_match_id)
);

-- Create indexes for foreign keys and performance
CREATE INDEX IF NOT EXISTS idx_afl_season_league_id ON afl.season(league_id);
CREATE INDEX IF NOT EXISTS idx_afl_round_season_id ON afl.round(season_id);
CREATE INDEX IF NOT EXISTS idx_afl_match_round_id ON afl.match(round_id);
CREATE INDEX IF NOT EXISTS idx_afl_match_home_club_match_id ON afl.match(home_club_match_id);
CREATE INDEX IF NOT EXISTS idx_afl_match_away_club_match_id ON afl.match(away_club_match_id);
CREATE INDEX IF NOT EXISTS idx_afl_club_season_club_id ON afl.club_season(club_id);
CREATE INDEX IF NOT EXISTS idx_afl_club_season_season_id ON afl.club_season(season_id);
CREATE INDEX IF NOT EXISTS idx_afl_club_match_match_id ON afl.club_match(match_id);
CREATE INDEX IF NOT EXISTS idx_afl_club_match_club_season_id ON afl.club_match(club_season_id);
CREATE INDEX IF NOT EXISTS idx_afl_player_club_id ON afl.player(club_id);
CREATE INDEX IF NOT EXISTS idx_afl_player_season_player_id ON afl.player_season(player_id);
CREATE INDEX IF NOT EXISTS idx_afl_player_season_club_season_id ON afl.player_season(club_season_id);
CREATE INDEX IF NOT EXISTS idx_afl_player_match_club_match_id ON afl.player_match(club_match_id);
CREATE INDEX IF NOT EXISTS idx_afl_player_match_player_season_id ON afl.player_match(player_season_id);

-- Create indexes for soft delete queries
CREATE INDEX IF NOT EXISTS idx_afl_league_deleted_at ON afl.league(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_season_deleted_at ON afl.season(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_round_deleted_at ON afl.round(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_match_deleted_at ON afl.match(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_club_deleted_at ON afl.club(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_club_season_deleted_at ON afl.club_season(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_club_match_deleted_at ON afl.club_match(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_player_deleted_at ON afl.player(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_player_season_deleted_at ON afl.player_season(deleted_at);
CREATE INDEX IF NOT EXISTS idx_afl_player_match_deleted_at ON afl.player_match(deleted_at);
