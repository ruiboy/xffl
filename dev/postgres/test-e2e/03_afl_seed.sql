-- AFL e2e test seed data (idempotent — safe to re-run)
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
SELECT l.id, 'AFL 2026'
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

-- Club seasons for Adelaide and Brisbane only
INSERT INTO afl.club_season (club_id, season_id)
SELECT c.id, s.id
FROM afl.club c, afl.season s
JOIN afl.league l ON s.league_id = l.id
WHERE c.name IN ('Adelaide Crows', 'Brisbane Lions') AND l.name = 'AFL' AND s.name = 'AFL 2026'
ON CONFLICT (club_id, season_id) DO NOTHING;

-- Players: 2 per team
INSERT INTO afl.player (name) VALUES
('Jordan Dawson'),
('Wayne Milera'),
('Henry Smith'),
('Hugh McCluggage'),
('Brock Thunder'),
('Kai Fernsby'),
('Lenny Voss'),
('Dax Morrow'),
('Theo Quillan'),
('Reid Calloway')
ON CONFLICT DO NOTHING;

-- Player seasons — Adelaide
INSERT INTO afl.player_season (player_id, club_season_id)
SELECT p.id, cs.id
FROM afl.player p, afl.club_season cs
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.season s ON cs.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
WHERE p.name IN ('Jordan Dawson', 'Wayne Milera')
  AND c.name = 'Adelaide Crows' AND l.name = 'AFL' AND s.name = 'AFL 2026'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

-- Player seasons — Brisbane
INSERT INTO afl.player_season (player_id, club_season_id)
SELECT p.id, cs.id
FROM afl.player p, afl.club_season cs
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.season s ON cs.season_id = s.id
JOIN afl.league l ON s.league_id = l.id
WHERE p.name IN ('Henry Smith', 'Hugh McCluggage')
  AND c.name = 'Brisbane Lions' AND l.name = 'AFL' AND s.name = 'AFL 2026'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

-- Round 1
INSERT INTO afl.round (season_id, name)
SELECT s.id, 'Round 1'
FROM afl.season s JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2026';

INSERT INTO afl.match (round_id, venue, start_dt, drv_result)
SELECT r.id, 'Adelaide Oval', '2025-03-15 14:10:00+09:30', 'no_result'
FROM afl.round r JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2026' AND r.name = 'Round 1';

INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT m.id, cs.id, 0, 0, 0
FROM afl.match m JOIN afl.round r ON m.round_id = r.id JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND r.name = 'Round 1' AND c.name IN ('Adelaide Crows', 'Brisbane Lions')
ON CONFLICT (club_season_id, match_id) DO NOTHING;

UPDATE afl.match SET
  home_club_match_id = (SELECT cm.id FROM afl.club_match cm JOIN afl.club_season cs ON cm.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id WHERE cm.match_id = afl.match.id AND c.name = 'Adelaide Crows'),
  away_club_match_id = (SELECT cm.id FROM afl.club_match cm JOIN afl.club_season cs ON cm.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id WHERE cm.match_id = afl.match.id AND c.name = 'Brisbane Lions')
WHERE venue = 'Adelaide Oval';

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 18, 12, 6, 0, 4, 2, 1
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Jordan Dawson' AND c.name = 'Adelaide Crows'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 22, 15, 4, 0, 6, 0, 2
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Wayne Milera' AND c.name = 'Adelaide Crows'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 20, 16, 5, 0, 8, 3, 2
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Henry Smith' AND c.name = 'Brisbane Lions'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 16, 10, 7, 0, 5, 1, 3
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id
WHERE p.name = 'Hugh McCluggage' AND c.name = 'Brisbane Lions'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

-- Round 2
INSERT INTO afl.round (season_id, name)
SELECT s.id, 'Round 2'
FROM afl.season s JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2026';

INSERT INTO afl.match (round_id, venue, start_dt, drv_result)
SELECT r.id, 'The Gabba', '2025-03-22 15:20:00+10:00', 'no_result'
FROM afl.round r JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2026' AND r.name = 'Round 2';

INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
SELECT m.id, cs.id, 0, 0, 0
FROM afl.match m JOIN afl.round r ON m.round_id = r.id JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id
JOIN afl.club_season cs ON cs.season_id = s.id JOIN afl.club c ON cs.club_id = c.id
WHERE l.name = 'AFL' AND r.name = 'Round 2' AND c.name IN ('Adelaide Crows', 'Brisbane Lions')
ON CONFLICT (club_season_id, match_id) DO NOTHING;

UPDATE afl.match SET
  home_club_match_id = (SELECT cm.id FROM afl.club_match cm JOIN afl.club_season cs ON cm.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id WHERE cm.match_id = afl.match.id AND c.name = 'Brisbane Lions'),
  away_club_match_id = (SELECT cm.id FROM afl.club_match cm JOIN afl.club_season cs ON cm.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id WHERE cm.match_id = afl.match.id AND c.name = 'Adelaide Crows')
WHERE venue = 'The Gabba';

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 15, 10, 8, 0, 3, 1, 2
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id)
WHERE p.name = 'Jordan Dawson' AND r.name = 'Round 2'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 25, 11, 5, 0, 8, 1, 0
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id)
WHERE p.name = 'Wayne Milera' AND r.name = 'Round 2'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 18, 14, 4, 0, 6, 2, 1
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id)
WHERE p.name = 'Henry Smith' AND r.name = 'Round 2'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 14, 8, 9, 0, 3, 0, 2
FROM afl.player_season ps JOIN afl.player p ON ps.player_id = p.id JOIN afl.club_season cs ON ps.club_season_id = cs.id JOIN afl.club c ON cs.club_id = c.id JOIN afl.club_match cm ON cm.club_season_id = cs.id JOIN afl.round r ON r.id = (SELECT round_id FROM afl.match WHERE id = cm.match_id)
WHERE p.name = 'Hugh McCluggage' AND r.name = 'Round 2'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

COMMIT;
