-- Create the ffl schema
CREATE SCHEMA IF NOT EXISTS ffl;

-- Create league table
CREATE TABLE IF NOT EXISTS ffl.league (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT uni_league_name UNIQUE (name)
);

-- Create season table
CREATE TABLE IF NOT EXISTS ffl.season (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    league_id INTEGER NOT NULL REFERENCES ffl.league(id) ON DELETE CASCADE
);

-- Create round table
CREATE TABLE IF NOT EXISTS ffl.round (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    season_id INTEGER NOT NULL REFERENCES ffl.season(id) ON DELETE CASCADE
);

-- Create match table
CREATE TABLE IF NOT EXISTS ffl.match (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    round_id INTEGER NOT NULL REFERENCES ffl.round(id) ON DELETE CASCADE,
    match_style VARCHAR(50) NOT NULL CHECK (match_style IN ('versus', 'bye', 'super_bye')),
    clubs JSONB,
    home_club_match_id INTEGER,
    away_club_match_id INTEGER,
    venue VARCHAR(255),
    start_dt TIMESTAMP WITH TIME ZONE,
    drv_result TEXT
);

-- Create club table
CREATE TABLE IF NOT EXISTS ffl.club (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT uni_club_name UNIQUE (name)
);

-- Create club_season table
CREATE TABLE IF NOT EXISTS ffl.club_season (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    club_id INTEGER NOT NULL REFERENCES ffl.club(id) ON DELETE CASCADE,
    season_id INTEGER NOT NULL REFERENCES ffl.season(id) ON DELETE CASCADE,
    drv_played INTEGER DEFAULT 0,
    drv_won INTEGER DEFAULT 0,
    drv_lost INTEGER DEFAULT 0,
    drv_drawn INTEGER DEFAULT 0,
    drv_for INTEGER DEFAULT 0,
    drv_against INTEGER DEFAULT 0,
    drv_extra_points INTEGER DEFAULT 0,
    drv_premiership_points INTEGER DEFAULT 0,
    CONSTRAINT uni_club_season UNIQUE (club_id, season_id)
);

-- Create club_match table
CREATE TABLE IF NOT EXISTS ffl.club_match (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    match_id INTEGER NOT NULL REFERENCES ffl.match(id) ON DELETE CASCADE,
    club_season_id INTEGER NOT NULL REFERENCES ffl.club_season(id) ON DELETE CASCADE,
    drv_score INTEGER DEFAULT 0,
    drv_premiership_points INTEGER DEFAULT 0
);

-- Add foreign key constraints for match table references to club_match
ALTER TABLE ffl.match 
ADD CONSTRAINT fk_match_home_club_match 
FOREIGN KEY (home_club_match_id) REFERENCES ffl.club_match(id);

ALTER TABLE ffl.match 
ADD CONSTRAINT fk_match_away_club_match 
FOREIGN KEY (away_club_match_id) REFERENCES ffl.club_match(id);

-- Create player table
CREATE TABLE IF NOT EXISTS ffl.player (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    club_id INTEGER REFERENCES ffl.club(id) ON DELETE CASCADE
);

-- Create player_season table
CREATE TABLE IF NOT EXISTS ffl.player_season (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    player_id INTEGER NOT NULL REFERENCES ffl.player(id) ON DELETE CASCADE,
    club_season_id INTEGER NOT NULL REFERENCES ffl.club_season(id) ON DELETE CASCADE,
    from_round_id INTEGER REFERENCES ffl.round(id),
    to_round_id INTEGER REFERENCES ffl.round(id),
    CONSTRAINT uni_player_season UNIQUE (player_id, club_season_id)
);

-- Create player_match table
CREATE TABLE IF NOT EXISTS ffl.player_match (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    club_match_id INTEGER NOT NULL REFERENCES ffl.club_match(id) ON DELETE CASCADE,
    player_season_id INTEGER NOT NULL REFERENCES ffl.player_season(id) ON DELETE CASCADE,
    position VARCHAR(255),
    interchange_positions TEXT,
    status VARCHAR(50) CHECK (status IN ('dnp', 'subbed_in')),
    score INTEGER DEFAULT 0
);

-- Create indexes for foreign keys and performance
CREATE INDEX IF NOT EXISTS idx_season_league_id ON ffl.season(league_id);
CREATE INDEX IF NOT EXISTS idx_round_season_id ON ffl.round(season_id);
CREATE INDEX IF NOT EXISTS idx_match_round_id ON ffl.match(round_id);
CREATE INDEX IF NOT EXISTS idx_match_home_club_match_id ON ffl.match(home_club_match_id);
CREATE INDEX IF NOT EXISTS idx_match_away_club_match_id ON ffl.match(away_club_match_id);
CREATE INDEX IF NOT EXISTS idx_club_season_club_id ON ffl.club_season(club_id);
CREATE INDEX IF NOT EXISTS idx_club_season_season_id ON ffl.club_season(season_id);
CREATE INDEX IF NOT EXISTS idx_club_match_match_id ON ffl.club_match(match_id);
CREATE INDEX IF NOT EXISTS idx_club_match_club_season_id ON ffl.club_match(club_season_id);
CREATE INDEX IF NOT EXISTS idx_player_club_id ON ffl.player(club_id);
CREATE INDEX IF NOT EXISTS idx_player_season_player_id ON ffl.player_season(player_id);
CREATE INDEX IF NOT EXISTS idx_player_season_club_season_id ON ffl.player_season(club_season_id);
CREATE INDEX IF NOT EXISTS idx_player_season_from_round_id ON ffl.player_season(from_round_id);
CREATE INDEX IF NOT EXISTS idx_player_season_to_round_id ON ffl.player_season(to_round_id);
CREATE INDEX IF NOT EXISTS idx_player_match_club_match_id ON ffl.player_match(club_match_id);
CREATE INDEX IF NOT EXISTS idx_player_match_player_season_id ON ffl.player_match(player_season_id);

-- Create indexes for soft delete queries
CREATE INDEX IF NOT EXISTS idx_league_deleted_at ON ffl.league(deleted_at);
CREATE INDEX IF NOT EXISTS idx_season_deleted_at ON ffl.season(deleted_at);
CREATE INDEX IF NOT EXISTS idx_round_deleted_at ON ffl.round(deleted_at);
CREATE INDEX IF NOT EXISTS idx_match_deleted_at ON ffl.match(deleted_at);
CREATE INDEX IF NOT EXISTS idx_club_deleted_at ON ffl.club(deleted_at);
CREATE INDEX IF NOT EXISTS idx_club_season_deleted_at ON ffl.club_season(deleted_at);
CREATE INDEX IF NOT EXISTS idx_club_match_deleted_at ON ffl.club_match(deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_deleted_at ON ffl.player(deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_season_deleted_at ON ffl.player_season(deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_match_deleted_at ON ffl.player_match(deleted_at);
