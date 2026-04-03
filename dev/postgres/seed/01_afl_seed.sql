-- AFL test data (idempotent — safe to re-run)
BEGIN;

-- Clear existing data (nullify match FKs first to break circular ref)
UPDATE afl.match SET home_club_match_id = NULL, away_club_match_id = NULL;
DELETE FROM afl.player_match;
DELETE FROM afl.club_match;
DELETE FROM afl.match;
DELETE FROM afl.round;
DELETE FROM afl.player_season;
DELETE FROM afl.club_season;
DELETE FROM afl.season;
DELETE FROM afl.player;

-- League (unique on name)
INSERT INTO afl.league (name) VALUES ('AFL') ON CONFLICT (name) DO NOTHING;

-- Season
INSERT INTO afl.season (league_id, name)
SELECT l.id, 'AFL 2025'
FROM afl.league l WHERE l.name = 'AFL'
ON CONFLICT DO NOTHING;

-- All 18 AFL clubs (unique on name)
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

-- Club seasons for Adelaide and Brisbane
INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
SELECT c.id, s.id, 12, 8, 4, 0, 1245, 1156, 32
FROM afl.club c, afl.season s
JOIN afl.league l ON s.league_id = l.id
WHERE c.name = 'Adelaide Crows' AND l.name = 'AFL' AND s.name = 'AFL 2025'
ON CONFLICT (club_id, season_id) DO NOTHING;

INSERT INTO afl.club_season (club_id, season_id, drv_played, drv_won, drv_lost, drv_drawn, drv_for, drv_against, drv_premiership_points)
SELECT c.id, s.id, 12, 7, 5, 0, 1198, 1167, 28
FROM afl.club c, afl.season s
JOIN afl.league l ON s.league_id = l.id
WHERE c.name = 'Brisbane Lions' AND l.name = 'AFL' AND s.name = 'AFL 2025'
ON CONFLICT (club_id, season_id) DO NOTHING;

-- Round 13
INSERT INTO afl.round (season_id, name)
SELECT s.id, 'Round 13'
FROM afl.season s
JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025';

-- Match: Adelaide v Brisbane at Adelaide Oval
INSERT INTO afl.match (round_id, venue, start_dt, drv_result)
SELECT r.id, 'Adelaide Oval', '2025-06-15 14:10:00+09:30', 'no_result'
FROM afl.round r
JOIN afl.season s ON r.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025' AND r.name = 'Round 13';

-- Home club match (Adelaide Crows)
INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT m.id, cs.id, 0, 0, 0
FROM afl.match m
JOIN afl.round r ON m.round_id = r.id
JOIN afl.season s ON r.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id
JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025' AND r.name = 'Round 13'
  AND c.name = 'Adelaide Crows'
ON CONFLICT (club_season_id, match_id) DO NOTHING;

-- Away club match (Brisbane Lions)
INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT m.id, cs.id, 0, 0, 0
FROM afl.match m
JOIN afl.round r ON m.round_id = r.id
JOIN afl.season s ON r.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id
JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025' AND r.name = 'Round 13'
  AND c.name = 'Brisbane Lions'
ON CONFLICT (club_season_id, match_id) DO NOTHING;

-- Link home/away club matches to the match
UPDATE afl.match SET
  home_club_match_id = (
    SELECT cm.id FROM afl.club_match cm
    JOIN afl.club_season cs ON cm.club_season_id = cs.id
    JOIN afl.club c ON cs.club_id = c.id
    WHERE cm.match_id = afl.match.id AND c.name = 'Adelaide Crows'
  ),
  away_club_match_id = (
    SELECT cm.id FROM afl.club_match cm
    JOIN afl.club_season cs ON cm.club_season_id = cs.id
    JOIN afl.club c ON cs.club_id = c.id
    WHERE cm.match_id = afl.match.id AND c.name = 'Brisbane Lions'
  )
WHERE afl.match.venue = 'Adelaide Oval';

-- Players
INSERT INTO afl.player (name) VALUES
('Jordan Dawson'),
('Rory Laird'),
('Ben Keays'),
('Lachie Neale'),
('Hugh McCluggage'),
('Dayne Zorko')
ON CONFLICT DO NOTHING;

-- Player seasons — Adelaide
INSERT INTO afl.player_season (player_id, club_season_id)
SELECT p.id, cs.id
FROM afl.player p, afl.club_season cs
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.season s ON cs.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
WHERE p.name IN ('Jordan Dawson', 'Rory Laird', 'Ben Keays')
  AND c.name = 'Adelaide Crows' AND l.name = 'AFL' AND s.name = 'AFL 2025'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

-- Player seasons — Brisbane
INSERT INTO afl.player_season (player_id, club_season_id)
SELECT p.id, cs.id
FROM afl.player p, afl.club_season cs
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.season s ON cs.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
WHERE p.name IN ('Lachie Neale', 'Hugh McCluggage', 'Dayne Zorko')
  AND c.name = 'Brisbane Lions' AND l.name = 'AFL' AND s.name = 'AFL 2025'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

-- Player match records — Adelaide players
INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 18, 12, 6, 0, 4, 2, 1
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Jordan Dawson' AND c.name = 'Adelaide Crows'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 22, 15, 4, 0, 6, 0, 2
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Rory Laird' AND c.name = 'Adelaide Crows'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 14, 18, 3, 0, 7, 1, 0
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Ben Keays' AND c.name = 'Adelaide Crows'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

-- Player match records — Brisbane players
INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 20, 16, 5, 0, 8, 3, 2
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Lachie Neale' AND c.name = 'Brisbane Lions'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 16, 10, 7, 0, 5, 1, 3
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Hugh McCluggage' AND c.name = 'Brisbane Lions'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 10, 14, 2, 0, 9, 0, 1
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Dayne Zorko' AND c.name = 'Brisbane Lions'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

-- Round 14
INSERT INTO afl.round (season_id, name)
SELECT s.id, 'Round 14'
FROM afl.season s JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025';

INSERT INTO afl.match (round_id, venue, start_dt, drv_result)
SELECT r.id, 'The Gabba', '2025-06-22 15:20:00+10:00', 'no_result'
FROM afl.round r JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025' AND r.name = 'Round 14';

INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT m.id, cs.id, 0, 0, 0
FROM afl.match m JOIN afl.round r ON m.round_id = r.id JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND r.name = 'Round 14' AND c.name = 'Adelaide Crows'
ON CONFLICT (club_season_id, match_id) DO NOTHING;

INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT m.id, cs.id, 0, 0, 0
FROM afl.match m JOIN afl.round r ON m.round_id = r.id JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND r.name = 'Round 14' AND c.name = 'Brisbane Lions'
ON CONFLICT (club_season_id, match_id) DO NOTHING;

UPDATE afl.match SET
  home_club_match_id = (SELECT cm.id FROM afl.club_match cm JOIN afl.club_season cs ON cm.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id WHERE cm.match_id = afl.match.id AND c.name = 'Brisbane Lions'),
  away_club_match_id = (SELECT cm.id FROM afl.club_match cm JOIN afl.club_season cs ON cm.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id WHERE cm.match_id = afl.match.id AND c.name = 'Adelaide Crows')
WHERE venue = 'The Gabba';

-- Round 14 player stats
INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 15, 10, 8, 0, 3, 1, 2
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Jordan Dawson' AND r.name = 'Round 14'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 25, 11, 5, 0, 8, 1, 0
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Rory Laird' AND r.name = 'Round 14'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 12, 20, 2, 0, 5, 0, 1
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Ben Keays' AND r.name = 'Round 14'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 18, 14, 4, 0, 6, 2, 1
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Lachie Neale' AND r.name = 'Round 14'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 14, 8, 9, 0, 3, 0, 2
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Hugh McCluggage' AND r.name = 'Round 14'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 8, 16, 3, 0, 11, 1, 0
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Dayne Zorko' AND r.name = 'Round 14'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

-- Round 15
INSERT INTO afl.round (season_id, name)
SELECT s.id, 'Round 15'
FROM afl.season s JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025';

INSERT INTO afl.match (round_id, venue, start_dt, drv_result)
SELECT r.id, 'MCG', '2025-06-29 14:10:00+10:00', 'no_result'
FROM afl.round r JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2025' AND r.name = 'Round 15';

INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT m.id, cs.id, 0, 0, 0
FROM afl.match m JOIN afl.round r ON m.round_id = r.id JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND r.name = 'Round 15' AND c.name = 'Adelaide Crows'
ON CONFLICT (club_season_id, match_id) DO NOTHING;

INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT m.id, cs.id, 0, 0, 0
FROM afl.match m JOIN afl.round r ON m.round_id = r.id JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND r.name = 'Round 15' AND c.name = 'Brisbane Lions'
ON CONFLICT (club_season_id, match_id) DO NOTHING;

UPDATE afl.match SET
  home_club_match_id = (SELECT cm.id FROM afl.club_match cm JOIN afl.club_season cs ON cm.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id WHERE cm.match_id = afl.match.id AND c.name = 'Adelaide Crows'),
  away_club_match_id = (SELECT cm.id FROM afl.club_match cm JOIN afl.club_season cs ON cm.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id WHERE cm.match_id = afl.match.id AND c.name = 'Brisbane Lions')
WHERE venue = 'MCG' AND round_id = (SELECT id FROM afl.round WHERE name = 'Round 15');

-- Round 15 player stats
INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 20, 14, 5, 0, 5, 3, 0
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Jordan Dawson' AND r.name = 'Round 15'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 19, 18, 3, 0, 4, 0, 1
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Rory Laird' AND r.name = 'Round 15'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 16, 15, 4, 0, 9, 2, 1
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Ben Keays' AND r.name = 'Round 15'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 22, 12, 6, 0, 7, 1, 3
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Lachie Neale' AND r.name = 'Round 15'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 18, 12, 6, 0, 7, 2, 1
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Hugh McCluggage' AND r.name = 'Round 15'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 12, 10, 4, 0, 6, 0, 2
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id) WHERE p.name = 'Dayne Zorko' AND r.name = 'Round 15'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

COMMIT;