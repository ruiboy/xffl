-- Insert AFL test data

-- Insert AFL League
INSERT INTO afl.league (name) VALUES ('AFL') ON CONFLICT (name) DO NOTHING;

-- Insert 2025 Season
INSERT INTO afl.season (league_id, name) 
SELECT l.id, 'AFL 2025'
FROM afl.league l WHERE l.name = 'AFL'
ON CONFLICT DO NOTHING;

-- Insert Round 13
INSERT INTO afl.round (season_id, name)
SELECT s.id, 'Round 13'
FROM afl.season s 
JOIN afl.league l ON s.league_id = l.id 
WHERE l.name = 'AFL' AND s.name = 'AFL 2025'
ON CONFLICT DO NOTHING;

-- Insert all 18 AFL clubs
INSERT INTO afl.club (name) VALUES
('Adelaide Crows'),
('Brisbane Lions'),
('Carlton Blues'),
('Collingwood Magpies'),
('Essendon Bombers'),
('Fremantle Dockers'),
('Geelong Cats'),
('Gold Coast Suns'),
('Greater Western Sydney Giants'),
('Hawthorn Hawks'),
('Melbourne Demons'),
('North Melbourne Kangaroos'),
('Port Adelaide Power'),
('Richmond Tigers'),
('St Kilda Saints'),
('Sydney Swans'),
('West Coast Eagles'),
('Western Bulldogs')
ON CONFLICT (name) DO NOTHING;

-- Insert club_season records for both teams
INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
SELECT 
    c.id,
    s.id,
    12, -- played 12 games so far
    8,  -- won 8 (Adelaide example)
    4,  -- lost 4
    0,  -- drawn 0
    1245, -- for score
    1156, -- against score  
    32  -- premiership points (8 wins * 4 points)
FROM afl.club c
JOIN afl.season s ON s.name = 'AFL 2025'
JOIN afl.league l ON s.league_id = l.id
WHERE c.name = 'Adelaide Crows' AND l.name = 'AFL'
ON CONFLICT (club_id, season_id) DO NOTHING;

INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
SELECT 
    c.id,
    s.id,
    12, -- played 12 games so far
    7,  -- won 7 (Brisbane example)
    5,  -- lost 5
    0,  -- drawn 0
    1198, -- for score
    1167, -- against score
    28  -- premiership points (7 wins * 4 points)
FROM afl.club c
JOIN afl.season s ON s.name = 'AFL 2025'
JOIN afl.league l ON s.league_id = l.id
WHERE c.name = 'Brisbane Lions' AND l.name = 'AFL'
ON CONFLICT (club_id, season_id) DO NOTHING;

-- First create the match record (without club references for now)
INSERT INTO afl.match (round_id, venue, start_dt, drv_result)
SELECT 
    r.id as round_id,
    'Adelaide Oval',
    '2025-06-15 14:10:00'::timestamp,
    'no_result'
FROM afl.round r
JOIN afl.season s ON r.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025' AND r.name = 'Round 13'
ON CONFLICT DO NOTHING;

-- Insert club_match records for both teams
-- Adelaide Crows (home team)
INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT 
    m.id,
    cs.id,
    0, -- score (match not played yet)
    0, -- premiership points (match not played yet)
    0  -- rushed behinds
FROM afl.match m
JOIN afl.round r ON m.round_id = r.id
JOIN afl.season s ON r.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id
JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025' AND r.name = 'Round 13' 
  AND c.name = 'Adelaide Crows' AND m.venue = 'Adelaide Oval'
ON CONFLICT DO NOTHING;

-- Brisbane Lions (away team)  
INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT 
    m.id,
    cs.id,
    0, -- score (match not played yet)
    0, -- premiership points (match not played yet)
    0  -- rushed behinds
FROM afl.match m
JOIN afl.round r ON m.round_id = r.id
JOIN afl.season s ON r.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id
JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025' AND r.name = 'Round 13' 
  AND c.name = 'Brisbane Lions' AND m.venue = 'Adelaide Oval'
ON CONFLICT DO NOTHING;

-- Insert Jordan Dawson player
INSERT INTO afl.player (name) VALUES ('Jordan Dawson') ON CONFLICT DO NOTHING;

-- Insert Jordan Dawson player_season (with Adelaide Crows for 2025)
INSERT INTO afl.player_season (player_id, club_season_id)
SELECT 
    p.id,
    cs.id
FROM afl.player p
JOIN afl.club_season cs ON cs.season_id = (
    SELECT s.id FROM afl.season s 
    JOIN afl.league l ON s.league_id = l.id 
    WHERE l.name = 'AFL' AND s.name = 'AFL 2025'
)
JOIN afl.club c ON cs.club_id = c.id
WHERE p.name = 'Jordan Dawson' AND c.name = 'Adelaide Crows'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

-- Insert Jordan Dawson player_match record for the Crows v Brisbane match
INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT 
    ps.id,
    cm.id,
    0, -- kicks (match not played yet)
    0, -- handballs  
    0, -- marks
    0, -- hitouts
    0, -- tackles
    0, -- goals
    0  -- behinds
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
JOIN afl.match m ON cm.match_id = m.id
JOIN afl.round r ON m.round_id = r.id
JOIN afl.season s ON r.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
WHERE p.name = 'Jordan Dawson' 
  AND c.name = 'Adelaide Crows'
  AND l.name = 'AFL' 
  AND s.name = 'AFL 2025' 
  AND r.name = 'Round 13'
  AND m.venue = 'Adelaide Oval'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;