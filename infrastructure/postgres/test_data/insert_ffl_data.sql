-- Insert test data for FFL schema

-- Insert league
INSERT INTO ffl.league (name) VALUES ('Premier Fantasy Football League')
ON CONFLICT (name) DO NOTHING;

-- Insert clubs
INSERT INTO ffl.club (name) VALUES
    ('Ruiboys'),
    ('The Howling Cows')
ON CONFLICT (name) DO NOTHING;

-- Insert season
INSERT INTO ffl.season (name, league_id) VALUES 
    ('2024 Season', (SELECT id FROM ffl.league WHERE name = 'Premier Fantasy Football League'));

-- Insert round
INSERT INTO ffl.round (name, season_id) VALUES 
    ('Round 1', (SELECT id FROM ffl.season WHERE name = '2024 Season'));

-- Insert club_season records
INSERT INTO ffl.club_season (club_id, season_id) VALUES
    ((SELECT id FROM ffl.club WHERE name = 'Ruiboys'), (SELECT id FROM ffl.season WHERE name = '2024 Season')),
    ((SELECT id FROM ffl.club WHERE name = 'The Howling Cows'), (SELECT id FROM ffl.season WHERE name = '2024 Season'));

-- Insert match (initially without club_match references)
INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES 
    ((SELECT id FROM ffl.round WHERE name = 'Round 1'), 
     'versus', 
     'MCG', 
     '2024-03-15 19:30:00+00');

-- Insert club_match records
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
        'home', jsonb_build_object('id', (SELECT c.id FROM ffl.club c WHERE c.name = 'Ruiboys'), 'name', 'Ruiboys'),
        'away', jsonb_build_object('id', (SELECT c.id FROM ffl.club c WHERE c.name = 'The Howling Cows'), 'name', 'The Howling Cows')
    ),
    drv_result = 'Ruiboys defeated The Howling Cows 85-72'
WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1');

-- Insert 30 players for Ruiboys
INSERT INTO ffl.player (name, club_id) VALUES
    ('Marcus Bontempelli', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Christian Petracca', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Lachie Neale', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Clayton Oliver', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Max Gawn', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Touk Miller', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Jack Steele', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Rory Laird', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Tim English', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Sam Walsh', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Jack Macrae', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Jeremy Cameron', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Tom Mitchell', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Darcy Parish', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Josh Dunkley', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Luke Ryan', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Nick Daicos', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Jordan Dawson', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Jayden Short', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Andrew Brayshaw', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Bailey Smith', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Zach Merrett', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Jake Lloyd', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Brodie Grundy', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Jack Crisp', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Hugh McCluggage', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Caleb Serong', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Errol Gulden', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Connor Rozee', (SELECT id FROM ffl.club WHERE name = 'Ruiboys')),
    ('Isaac Heeney', (SELECT id FROM ffl.club WHERE name = 'Ruiboys'));

-- Insert 10 players for The Howling Cows
INSERT INTO ffl.player (name, club_id) VALUES
    ('Dustin Martin', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Patrick Cripps', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Lance Franklin', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Travis Boak', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Nat Fyfe', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Tom Hawkins', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Elliot Yeo', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Scott Pendlebury', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Robbie Gray', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')),
    ('Joel Selwood', (SELECT id FROM ffl.club WHERE name = 'The Howling Cows'));

-- Insert player_season records for all players
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id)
SELECT 
    p.id,
    cs.id,
    r.id
FROM ffl.player p
JOIN ffl.club c ON p.club_id = c.id
JOIN ffl.club_season cs ON cs.club_id = c.id
JOIN ffl.season s ON cs.season_id = s.id
JOIN ffl.round r ON r.season_id = s.id
WHERE s.name = '2024 Season' AND r.name = 'Round 1';

-- Insert player_match records for Ruiboys players (22 starting players)
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, score)
SELECT 
    cm.id,
    ps.id,
    CASE 
        WHEN ROW_NUMBER() OVER (ORDER BY ps.id) <= 6 THEN 'DEF'
        WHEN ROW_NUMBER() OVER (ORDER BY ps.id) <= 14 THEN 'MID'
        WHEN ROW_NUMBER() OVER (ORDER BY ps.id) <= 20 THEN 'FWD'
        ELSE 'RUC'
    END,
    FLOOR(RANDOM() * 100 + 50)::INTEGER
FROM ffl.player_season ps
JOIN ffl.club_season cs ON ps.club_season_id = cs.id
JOIN ffl.club_match cm ON cm.club_season_id = cs.id
JOIN ffl.club c ON cs.club_id = c.id
WHERE c.name = 'Ruiboys'
ORDER BY ps.id
LIMIT 22;

-- Insert player_match records for The Howling Cows players (all 10 players)
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, score)
SELECT 
    cm.id,
    ps.id,
    CASE 
        WHEN ROW_NUMBER() OVER (ORDER BY ps.id) <= 3 THEN 'DEF'
        WHEN ROW_NUMBER() OVER (ORDER BY ps.id) <= 7 THEN 'MID'
        WHEN ROW_NUMBER() OVER (ORDER BY ps.id) <= 9 THEN 'FWD'
        ELSE 'RUC'
    END,
    FLOOR(RANDOM() * 100 + 40)::INTEGER
FROM ffl.player_season ps
JOIN ffl.club_season cs ON ps.club_season_id = cs.id
JOIN ffl.club_match cm ON cm.club_season_id = cs.id
JOIN ffl.club c ON cs.club_id = c.id
WHERE c.name = 'The Howling Cows'
ORDER BY ps.id;

-- Update club_season statistics based on match results
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