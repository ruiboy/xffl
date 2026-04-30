-- FFL e2e test seed data (idempotent — safe to re-run)
BEGIN;

-- Clear existing data and reset identity sequences so re-runs produce stable IDs.
-- CASCADE handles circular FKs between match and club_match.
TRUNCATE TABLE
    ffl.player_match,
    ffl.player_season,
    ffl.club_match,
    ffl.match,
    ffl.round,
    ffl.club_season,
    ffl.season,
    ffl.player
RESTART IDENTITY CASCADE;

-- League (unique on name)
INSERT INTO ffl.league (name) VALUES ('FFL')
ON CONFLICT (name) DO NOTHING;

-- Clubs (unique on name)
INSERT INTO ffl.club (name) VALUES
    ('Ruiboys'),
    ('The Howling Cows')
ON CONFLICT (name) DO NOTHING;

-- Season
INSERT INTO ffl.season (name, league_id, afl_season_id) VALUES
    ('FFL 2026', (SELECT id FROM ffl.league WHERE name = 'FFL'),
     (SELECT id FROM afl.season WHERE name = 'AFL 2026'));

-- Club seasons
INSERT INTO ffl.club_season (club_id, season_id) VALUES
    ((SELECT id FROM ffl.club WHERE name = 'Ruiboys'), (SELECT id FROM ffl.season WHERE name = 'FFL 2026')),
    ((SELECT id FROM ffl.club WHERE name = 'The Howling Cows'),     (SELECT id FROM ffl.season WHERE name = 'FFL 2026'));

-- Players: Ruiboys use Adelaide players, The Howling Cows use Brisbane players
INSERT INTO ffl.player (afl_player_id, drv_name)
SELECT ap.id, ap.name FROM afl.player ap WHERE ap.name IN ('Jordan Dawson', 'Wayne Milera');

INSERT INTO ffl.player (afl_player_id, drv_name)
SELECT ap.id, ap.name FROM afl.player ap WHERE ap.name IN (
    'Henry Smith', 'Hugh McCluggage',
    'Brock Thunder', 'Kai Fernsby', 'Lenny Voss', 'Dax Morrow', 'Theo Quillan', 'Reid Calloway'
);

-- Round 1
INSERT INTO ffl.round (name, season_id, afl_round_id)
SELECT 'Round 1', s.id, ar.id
FROM ffl.season s, afl.round ar
JOIN afl.season as2 ON ar.season_id = as2.id
JOIN afl.league al ON as2.league_id = al.id
WHERE s.name = 'FFL 2026' AND al.name = 'AFL' AND as2.name = 'AFL 2026' AND ar.name = 'Round 1';

-- Player seasons — Ruiboys
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id)
SELECT p.id, cs.id, r.id
FROM ffl.player p
JOIN afl.player ap ON p.afl_player_id = ap.id
JOIN ffl.club_season cs ON cs.club_id = (SELECT id FROM ffl.club WHERE name = 'Ruiboys')
JOIN ffl.round r ON r.season_id = cs.season_id
WHERE r.name = 'Round 1' AND ap.name IN ('Jordan Dawson', 'Wayne Milera');

-- Player seasons — The Howling Cows (Henry Smith + Hugh McCluggage assigned to positions;
-- remaining 6 are squad-only with no player_match, available in the team builder)
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id)
SELECT p.id, cs.id, r.id
FROM ffl.player p
JOIN afl.player ap ON p.afl_player_id = ap.id
JOIN ffl.club_season cs ON cs.club_id = (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')
JOIN ffl.round r ON r.season_id = cs.season_id
WHERE r.name = 'Round 1' AND ap.name IN (
    'Henry Smith', 'Hugh McCluggage',
    'Brock Thunder', 'Kai Fernsby', 'Lenny Voss', 'Dax Morrow', 'Theo Quillan', 'Reid Calloway'
);

-- Link FFL player seasons to their AFL player seasons (same player, AFL 2026 season)
UPDATE ffl.player_season fps
SET afl_player_season_id = aps.id
FROM afl.player_season aps
JOIN afl.player ap ON aps.player_id = ap.id
JOIN afl.club_season acs ON aps.club_season_id = acs.id
JOIN afl.season as2 ON acs.season_id = as2.id
JOIN ffl.player fp ON fp.afl_player_id = ap.id
WHERE fps.player_id = fp.id
  AND as2.name = 'AFL 2026';

-- Round 1 match
INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES
    ((SELECT id FROM ffl.round WHERE name = 'Round 1'), 'versus', 'MCG', '2025-03-15 19:30:00+00');

INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     85, 4),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     72, 0);

UPDATE ffl.match SET
    home_club_match_id = (SELECT cm.id FROM ffl.club_match cm JOIN ffl.club_season cs ON cm.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
    away_club_match_id = (SELECT cm.id FROM ffl.club_match cm JOIN ffl.club_season cs ON cm.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
    drv_result = 'Ruiboys defeated The Howling Cows 85-72'
WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1');

-- Round 1 player matches; 1 player in team — Ruiboys
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_score)
SELECT cm.id, ps.id, 'goals', 'played', 42
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.round r ON r.id = (SELECT round_id FROM ffl.match WHERE id = cm.match_id)
WHERE c.name = 'Ruiboys' AND ap.name = 'Jordan Dawson' AND r.name = 'Round 1';

-- Round 1 player matches; 1 player in team — The Howling Cows
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_score)
SELECT cm.id, ps.id, 'goals', 'played', 38
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.round r ON r.id = (SELECT round_id FROM ffl.match WHERE id = cm.match_id)
WHERE c.name = 'The Howling Cows' AND ap.name = 'Henry Smith' AND r.name = 'Round 1';

-- Club season stats after Round 1
UPDATE ffl.club_season SET drv_played = 1, drv_won = 1, drv_lost = 0, drv_for = 85, drv_against = 72, drv_premiership_points = 4
WHERE id = (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys');

UPDATE ffl.club_season SET drv_played = 1, drv_won = 0, drv_lost = 1, drv_for = 72, drv_against = 85, drv_premiership_points = 0
WHERE id = (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows');

-- Round 2
INSERT INTO ffl.round (name, season_id, afl_round_id)
SELECT 'Round 2', s.id, ar.id
FROM ffl.season s, afl.round ar
JOIN afl.season as2 ON ar.season_id = as2.id
JOIN afl.league al ON as2.league_id = al.id
WHERE s.name = 'FFL 2026' AND al.name = 'AFL' AND as2.name = 'AFL 2026' AND ar.name = 'Round 2';

INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES
    ((SELECT id FROM ffl.round WHERE name = 'Round 2'), 'versus', 'MCG', '2025-03-22 19:30:00+00');

INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 2')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     68, 0),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 2')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     91, 4);

UPDATE ffl.match SET
    home_club_match_id = (SELECT cm.id FROM ffl.club_match cm JOIN ffl.club_season cs ON cm.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys' AND cm.match_id = ffl.match.id),
    away_club_match_id = (SELECT cm.id FROM ffl.club_match cm JOIN ffl.club_season cs ON cm.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows' AND cm.match_id = ffl.match.id),
    drv_result = 'The Howling Cows defeated Ruiboys 91-68'
WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 2');

-- Round 2 player matches; 1 player in team — Ruiboys
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_score)
SELECT cm.id, ps.id, 'goals', 'played', 31
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.round r ON r.id = (SELECT round_id FROM ffl.match WHERE id = cm.match_id)
WHERE c.name = 'Ruiboys' AND ap.name = 'Jordan Dawson' AND r.name = 'Round 2';

-- Round 2 player matches; 1 player in team  — The Howling Cows
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_score)
SELECT cm.id, ps.id, 'goals', 'played', 48
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.round r ON r.id = (SELECT round_id FROM ffl.match WHERE id = cm.match_id)
WHERE c.name = 'The Howling Cows' AND ap.name = 'Henry Smith' AND r.name = 'Round 2';

-- Club season stats after Round 2
UPDATE ffl.club_season SET drv_played = 2, drv_won = 1, drv_lost = 1, drv_for = 153, drv_against = 163, drv_premiership_points = 4
WHERE id = (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys');

UPDATE ffl.club_season SET drv_played = 2, drv_won = 1, drv_lost = 1, drv_for = 163, drv_against = 153, drv_premiership_points = 4
WHERE id = (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows');

-- Round 3 — linked to AFL Round 3 for live-round mapping tests
INSERT INTO ffl.round (name, season_id, afl_round_id)
SELECT 'Round 3', s.id, ar.id
FROM ffl.season s, afl.round ar
JOIN afl.season as2 ON ar.season_id = as2.id
JOIN afl.league al ON as2.league_id = al.id
WHERE s.name = 'FFL 2026' AND al.name = 'AFL' AND as2.name = 'AFL 2026' AND ar.name = 'Round 3';

INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES
    ((SELECT id FROM ffl.round WHERE name = 'Round 3'), 'versus', 'MCG', '2026-01-15 14:10:00+10:30');

INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 3')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     0, 0),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 3')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     0, 0);

UPDATE ffl.match SET
    home_club_match_id = (SELECT cm.id FROM ffl.club_match cm JOIN ffl.club_season cs ON cm.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys' AND cm.match_id = ffl.match.id),
    away_club_match_id = (SELECT cm.id FROM ffl.club_match cm JOIN ffl.club_season cs ON cm.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows' AND cm.match_id = ffl.match.id)
WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 3');

-- Link FFL player matches to AFL player matches via round bridge + shared player
UPDATE ffl.player_match fpm
SET afl_player_match_id = apm.id
FROM afl.player_match apm
JOIN afl.player_season aps ON apm.player_season_id = aps.id
JOIN afl.player ap ON aps.player_id = ap.id
JOIN afl.club_match acm ON apm.club_match_id = acm.id
JOIN afl.match am ON acm.match_id = am.id,
ffl.player_season fps,
ffl.player fp,
ffl.club_match fcm,
ffl.match fm,
ffl.round fr
WHERE fpm.player_season_id = fps.id
  AND fps.player_id = fp.id
  AND fp.afl_player_id = ap.id
  AND fpm.club_match_id = fcm.id
  AND fcm.match_id = fm.id
  AND fm.round_id = fr.id
  AND fr.afl_round_id = am.round_id;

COMMIT;
