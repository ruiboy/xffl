-- Create the ffl schema
CREATE SCHEMA IF NOT EXISTS ffl;

-- Create club table
CREATE TABLE IF NOT EXISTS ffl.club (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT uni_club_name UNIQUE (name)
);

-- Create player table
CREATE TABLE IF NOT EXISTS ffl.player (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    club_id INTEGER REFERENCES ffl.club(id) ON DELETE CASCADE
);

-- Create index on club_id for better query performance
CREATE INDEX IF NOT EXISTS idx_player_club_id ON ffl.player(club_id);

-- Create index on deleted_at for soft delete queries
CREATE INDEX IF NOT EXISTS idx_club_deleted_at ON ffl.club(deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_deleted_at ON ffl.player(deleted_at); 