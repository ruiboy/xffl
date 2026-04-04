-- FFL test data (idempotent — safe to re-run)
BEGIN;

-- Clear existing data (nullify match FKs first to break circular ref)
UPDATE ffl.match SET home_club_match_id = NULL, away_club_match_id = NULL;
DELETE FROM ffl.player_match;
DELETE FROM ffl.player_season;
DELETE FROM ffl.club_match;
DELETE FROM ffl.match;
DELETE FROM ffl.round;
DELETE FROM ffl.club_season;
DELETE FROM ffl.season;
DELETE FROM ffl.player;

-- League (unique on name)
INSERT INTO ffl.league (name) VALUES ('Premier Fantasy Football League')
ON CONFLICT (name) DO NOTHING;

-- Clubs (unique on name)
INSERT INTO ffl.club (name) VALUES
    ('Ruiboys'),
    ('The Howling Cows')
ON CONFLICT (name) DO NOTHING;

-- Season
INSERT INTO ffl.season (name, league_id) VALUES
    ('2024 Season', (SELECT id FROM ffl.league WHERE name = 'Premier Fantasy Football League'));

-- Round
INSERT INTO ffl.round (name, season_id) VALUES
    ('Round 1', (SELECT id FROM ffl.season WHERE name = '2024 Season'));

-- Club seasons
INSERT INTO ffl.club_season (club_id, season_id) VALUES
    ((SELECT id FROM ffl.club WHERE name = 'Ruiboys'), (SELECT id FROM ffl.season WHERE name = '2024 Season')),
    ((SELECT id FROM ffl.club WHERE name = 'The Howling Cows'), (SELECT id FROM ffl.season WHERE name = '2024 Season'));

-- Match (initially without club_match references)
INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES
    ((SELECT id FROM ffl.round WHERE name = 'Round 1'),
     'versus',
     'MCG',
     '2024-03-15 19:30:00+00');

-- Club match records
INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     85, 4),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     72, 0);

-- Update match with club_match references
UPDATE ffl.match
SET home_club_match_id = (SELECT cm.id FROM ffl.club_match cm JOIN ffl.club_season cs ON cm.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
    away_club_match_id = (SELECT cm.id FROM ffl.club_match cm JOIN ffl.club_season cs ON cm.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
    clubs = jsonb_build_object(
        'home', jsonb_build_object('club_season_id', (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys')),
        'away', jsonb_build_object('club_season_id', (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'))
    ),
    drv_result = 'Ruiboys defeated The Howling Cows 85-72'
WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1');

-- 30 players for Ruiboys (afl_player_id looked up by name, drv_name derived)
INSERT INTO ffl.player (afl_player_id, drv_name)
SELECT ap.id, ap.name FROM afl.player ap WHERE ap.name IN (
    'Marcus Bontempelli', 'Christian Petracca', 'Lachie Neale', 'Clayton Oliver',
    'Max Gawn', 'Touk Miller', 'Jack Steele', 'Rory Laird', 'Tim English',
    'Sam Walsh', 'Jack Macrae', 'Jeremy Cameron', 'Tom Mitchell', 'Darcy Parish',
    'Josh Dunkley', 'Luke Ryan', 'Nick Daicos', 'Jordan Dawson', 'Jayden Short',
    'Andrew Brayshaw', 'Bailey Smith', 'Zach Merrett', 'Jake Lloyd', 'Brodie Grundy',
    'Jack Crisp', 'Hugh McCluggage', 'Caleb Serong', 'Errol Gulden', 'Connor Rozee',
    'Isaac Heeney'
);

-- 10 players for The Howling Cows
INSERT INTO ffl.player (afl_player_id, drv_name)
SELECT ap.id, ap.name FROM afl.player ap WHERE ap.name IN (
    'Dustin Martin', 'Patrick Cripps', 'Lance Franklin', 'Travis Boak', 'Nat Fyfe',
    'Tom Hawkins', 'Elliot Yeo', 'Scott Pendlebury', 'Robbie Gray', 'Joel Selwood'
);

-- Player season records — Ruiboys
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id)
SELECT
    p.id,
    cs.id,
    r.id
FROM ffl.player p
JOIN afl.player ap ON p.afl_player_id = ap.id
JOIN ffl.club_season cs ON cs.club_id = (SELECT id FROM ffl.club WHERE name = 'Ruiboys')
JOIN ffl.season s ON cs.season_id = s.id
JOIN ffl.round r ON r.season_id = s.id
WHERE s.name = '2024 Season' AND r.name = 'Round 1'
  AND ap.name IN (
    'Marcus Bontempelli', 'Christian Petracca', 'Lachie Neale', 'Clayton Oliver',
    'Max Gawn', 'Touk Miller', 'Jack Steele', 'Rory Laird', 'Tim English',
    'Sam Walsh', 'Jack Macrae', 'Jeremy Cameron', 'Tom Mitchell', 'Darcy Parish',
    'Josh Dunkley', 'Luke Ryan', 'Nick Daicos', 'Jordan Dawson', 'Jayden Short',
    'Andrew Brayshaw', 'Bailey Smith', 'Zach Merrett', 'Jake Lloyd', 'Brodie Grundy',
    'Jack Crisp', 'Hugh McCluggage', 'Caleb Serong', 'Errol Gulden', 'Connor Rozee',
    'Isaac Heeney'
  );

-- Player season records — The Howling Cows
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id)
SELECT
    p.id,
    cs.id,
    r.id
FROM ffl.player p
JOIN afl.player ap ON p.afl_player_id = ap.id
JOIN ffl.club_season cs ON cs.club_id = (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')
JOIN ffl.season s ON cs.season_id = s.id
JOIN ffl.round r ON r.season_id = s.id
WHERE s.name = '2024 Season' AND r.name = 'Round 1'
  AND ap.name IN (
    'Dustin Martin', 'Patrick Cripps', 'Lance Franklin', 'Travis Boak', 'Nat Fyfe',
    'Tom Hawkins', 'Elliot Yeo', 'Scott Pendlebury', 'Robbie Gray', 'Joel Selwood'
  );

-- Player match records for Ruiboys (7 starters + 2 bench)
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_score)
SELECT
    cm.id,
    ps.id,
    CASE ROW_NUMBER() OVER (ORDER BY ps.id)
        WHEN 1 THEN 'goals'
        WHEN 2 THEN 'kicks'
        WHEN 3 THEN 'handballs'
        WHEN 4 THEN 'marks'
        WHEN 5 THEN 'tackles'
        WHEN 6 THEN 'hitouts'
        WHEN 7 THEN 'star'
    END,
    'played',
    FLOOR(RANDOM() * 30 + 5)::INTEGER
FROM ffl.player_season ps
JOIN ffl.club_season cs ON ps.club_season_id = cs.id
JOIN ffl.club_match cm ON cm.club_season_id = cs.id
JOIN ffl.club c ON cs.club_id = c.id
WHERE c.name = 'Ruiboys'
ORDER BY ps.id
LIMIT 7;

-- Bench players for Ruiboys
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, backup_positions, interchange_position, drv_score)
SELECT
    cm.id,
    ps.id,
    CASE ROW_NUMBER() OVER (ORDER BY ps.id)
        WHEN 1 THEN 'goals'
        WHEN 2 THEN 'kicks'
    END,
    'played',
    CASE ROW_NUMBER() OVER (ORDER BY ps.id)
        WHEN 1 THEN 'goals,kicks'
        WHEN 2 THEN NULL
    END,
    CASE ROW_NUMBER() OVER (ORDER BY ps.id)
        WHEN 1 THEN NULL
        WHEN 2 THEN 'kicks'
    END,
    FLOOR(RANDOM() * 20 + 1)::INTEGER
FROM ffl.player_season ps
JOIN ffl.club_season cs ON ps.club_season_id = cs.id
JOIN ffl.club_match cm ON cm.club_season_id = cs.id
JOIN ffl.club c ON cs.club_id = c.id
WHERE c.name = 'Ruiboys'
ORDER BY ps.id
OFFSET 7 LIMIT 2;

-- Player match records for The Howling Cows (7 starters)
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_score)
SELECT
    cm.id,
    ps.id,
    CASE ROW_NUMBER() OVER (ORDER BY ps.id)
        WHEN 1 THEN 'goals'
        WHEN 2 THEN 'kicks'
        WHEN 3 THEN 'handballs'
        WHEN 4 THEN 'marks'
        WHEN 5 THEN 'tackles'
        WHEN 6 THEN 'hitouts'
        WHEN 7 THEN 'star'
    END,
    'played',
    FLOOR(RANDOM() * 30 + 5)::INTEGER
FROM ffl.player_season ps
JOIN ffl.club_season cs ON ps.club_season_id = cs.id
JOIN ffl.club_match cm ON cm.club_season_id = cs.id
JOIN ffl.club c ON cs.club_id = c.id
WHERE c.name = 'The Howling Cows'
ORDER BY ps.id
LIMIT 7;

-- Update club_season statistics
UPDATE ffl.club_season
SET drv_played = 1,
    drv_won = 1,
    drv_lost = 0,
    drv_drawn = 0,
    drv_for = 85,
    drv_against = 72,
    drv_premiership_points = 4
WHERE id IN (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys');

UPDATE ffl.club_season
SET drv_played = 1,
    drv_won = 0,
    drv_lost = 1,
    drv_drawn = 0,
    drv_for = 72,
    drv_against = 85,
    drv_premiership_points = 0
WHERE id IN (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows');

COMMIT;
