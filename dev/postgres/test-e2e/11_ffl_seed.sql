-- FFL e2e test seed data (idempotent — safe to re-run)
BEGIN;

-- Clear existing data and reset identity sequences so re-runs produce stable IDs.
-- CASCADE handles referential dependencies.
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

-- Players: Ruiboys use Adelaide players, The Howling Cows use Brisbane players.
INSERT INTO ffl.player (afl_player_id)
SELECT ap.id FROM afl.player ap WHERE ap.name IN ('Jordan Dawson', 'Wayne Milera');

INSERT INTO ffl.player (afl_player_id)
SELECT ap.id FROM afl.player ap WHERE ap.name IN (
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

-- Player seasons — Ruiboys (linked to AFL player_season in a single INSERT)
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT fp.id, cs.id, r.id, aps.id
FROM ffl.player fp
JOIN afl.player ap ON fp.afl_player_id = ap.id
JOIN afl.player_season aps ON aps.player_id = ap.id
JOIN afl.club_season acs ON aps.club_season_id = acs.id
JOIN afl.season as2 ON acs.season_id = as2.id
JOIN ffl.club_season cs ON cs.club_id = (SELECT id FROM ffl.club WHERE name = 'Ruiboys')
JOIN ffl.round r ON r.season_id = cs.season_id AND r.name = 'Round 1'
WHERE ap.name IN ('Jordan Dawson', 'Wayne Milera') AND as2.name = 'AFL 2026';

-- Player seasons — The Howling Cows (Henry Smith + Hugh McCluggage assigned to positions;
-- remaining 6 are squad-only with no player_match, available in the team builder)
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT fp.id, cs.id, r.id, aps.id
FROM ffl.player fp
JOIN afl.player ap ON fp.afl_player_id = ap.id
JOIN afl.player_season aps ON aps.player_id = ap.id
JOIN afl.club_season acs ON aps.club_season_id = acs.id
JOIN afl.season as2 ON acs.season_id = as2.id
JOIN ffl.club_season cs ON cs.club_id = (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')
JOIN ffl.round r ON r.season_id = cs.season_id AND r.name = 'Round 1'
WHERE ap.name IN (
    'Henry Smith', 'Hugh McCluggage',
    'Brock Thunder', 'Kai Fernsby', 'Lenny Voss', 'Dax Morrow', 'Theo Quillan', 'Reid Calloway'
) AND as2.name = 'AFL 2026';

-- Round 1 match
INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES
    ((SELECT id FROM ffl.round WHERE name = 'Round 1'), 'versus', 'MCG', '2025-03-15 19:30:00+00');

INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, side) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     85, 4, 'home'),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     72, 0, 'away');

UPDATE ffl.match SET
    drv_result = 'Ruiboys defeated The Howling Cows 85-72'
WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 1');

-- Round 1 player matches; 1 player in team — Ruiboys
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'goals', 'named', 'played', 42
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.round r ON r.id = (SELECT round_id FROM ffl.match WHERE id = cm.match_id)
WHERE c.name = 'Ruiboys' AND ap.name = 'Jordan Dawson' AND r.name = 'Round 1';

-- Round 1 player matches; 1 player in team — The Howling Cows
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'goals', 'named', 'played', 38
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

INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, side) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 2')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     68, 0, 'home'),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 2')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     91, 4, 'away');

UPDATE ffl.match SET
    drv_result = 'The Howling Cows defeated Ruiboys 91-68'
WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 2');

-- Round 2 player matches; 1 player in team — Ruiboys
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'goals', 'named', 'played', 31
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.round r ON r.id = (SELECT round_id FROM ffl.match WHERE id = cm.match_id)
WHERE c.name = 'Ruiboys' AND ap.name = 'Jordan Dawson' AND r.name = 'Round 2';

-- Round 2 player matches — The Howling Cows
-- Henry Smith: goals starter, played (used for subs-mode test; aflMatchStarted = true)
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'goals', 'named', 'played', 48
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.match fm ON cm.match_id = fm.id JOIN ffl.round r ON fm.round_id = r.id
WHERE c.name = 'The Howling Cows' AND ap.name = 'Henry Smith' AND r.name = 'Round 2';

-- Hugh McCluggage: kicks starter, DNP — candidate to be subbed out
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'kicks', 'named', 'dnp', 0
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.match fm ON cm.match_id = fm.id JOIN ffl.round r ON fm.round_id = r.id
WHERE c.name = 'The Howling Cows' AND ap.name = 'Hugh McCluggage' AND r.name = 'Round 2';

-- Brock Thunder: bench covering kicks, played — will sub in for Hugh
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, backup_positions, drv_score)
SELECT cm.id, ps.id, 'kicks', 'named', 'played', 'kicks', 20
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.match fm ON cm.match_id = fm.id JOIN ffl.round r ON fm.round_id = r.id
WHERE c.name = 'The Howling Cows' AND ap.name = 'Brock Thunder' AND r.name = 'Round 2';

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

INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, side) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 3')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     0, 0, 'home'),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 3')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     0, 0, 'away');

-- Round 4 — linked to AFL Round 4; used for the interchange e2e test scenario.
-- The Howling Cows club_match for Round 4 = id 8 (insert order: R1-R=1, R1-C=2, R2-R=3, R2-C=4,
--   R3-R=5, R3-C=6, R4-R=7, R4-C=8).
INSERT INTO ffl.round (name, season_id, afl_round_id)
VALUES (
    'Round 4',
    (SELECT id FROM ffl.season WHERE name = 'FFL 2026'),
    (SELECT r.id FROM afl.round r JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id WHERE l.name = 'AFL' AND s.name = 'AFL 2026' AND r.name = 'Round 4')
);

INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES
    ((SELECT id FROM ffl.round WHERE name = 'Round 4'), 'versus', 'MCG', '2026-01-22 19:30:00+00');

INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, side) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 4')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     0, 0, 'home'),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 4')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     0, 0, 'away');

-- Round 4 player matches — The Howling Cows (interchange test scenario)
-- Henry Smith: goals starter, played — will remain (higher score)
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'goals', 'named', 'played', 48
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.match fm ON cm.match_id = fm.id JOIN ffl.round r ON fm.round_id = r.id
WHERE c.name = 'The Howling Cows' AND ap.name = 'Henry Smith' AND r.name = 'Round 4';

-- Hugh McCluggage: goals starter, played, lower score — will be displaced by interchange
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'goals', 'named', 'played', 30
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.match fm ON cm.match_id = fm.id JOIN ffl.round r ON fm.round_id = r.id
WHERE c.name = 'The Howling Cows' AND ap.name = 'Hugh McCluggage' AND r.name = 'Round 4';

-- Brock Thunder: bench covering goals with interchange slot, played, score 50 — will interchange in
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, backup_positions, interchange_position, drv_score)
SELECT cm.id, ps.id, NULL, 'named', 'played', 'goals', 'goals', 50
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.match fm ON cm.match_id = fm.id JOIN ffl.round r ON fm.round_id = r.id
WHERE c.name = 'The Howling Cows' AND ap.name = 'Brock Thunder' AND r.name = 'Round 4';

-- Round 5 — Brisbane Lions have an AFL bye; The Howling Cows players score via season average.
-- FFL club_match insert order: R1-Rui=1, R1-Cows=2, R2-Rui=3, R2-Cows=4,
--   R3-Rui=5, R3-Cows=6, R4-Rui=7, R4-Cows=8, R5-Rui=9, R5-Cows=10
INSERT INTO ffl.round (name, season_id, afl_round_id)
VALUES (
    'Round 5',
    (SELECT id FROM ffl.season WHERE name = 'FFL 2026'),
    (SELECT r.id FROM afl.round r JOIN afl.season s ON r.season_id = s.id JOIN afl.league l ON s.league_id = l.id WHERE l.name = 'AFL' AND s.name = 'AFL 2026' AND r.name = 'Round 5')
);

INSERT INTO ffl.match (round_id, match_style, venue, start_dt) VALUES
    ((SELECT id FROM ffl.round WHERE name = 'Round 5'), 'versus', 'MCG', '2026-01-29 19:30:00+00');

INSERT INTO ffl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, side) VALUES
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 5')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'Ruiboys'),
     0, 0, 'home'),
    ((SELECT id FROM ffl.match WHERE round_id = (SELECT id FROM ffl.round WHERE name = 'Round 5')),
     (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = 'The Howling Cows'),
     0, 0, 'away');

-- Henry Smith: goals starter, bye — scores 38 (pre-computed from season average)
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'goals', 'named', 'bye', 38
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.match fm ON cm.match_id = fm.id JOIN ffl.round r ON fm.round_id = r.id
WHERE c.name = 'The Howling Cows' AND ap.name = 'Henry Smith' AND r.name = 'Round 5';

-- Hugh McCluggage: kicks starter, bye — scores 14 (pre-computed from season average)
INSERT INTO ffl.player_match (club_match_id, player_season_id, position, status, drv_afl_status, drv_score)
SELECT cm.id, ps.id, 'kicks', 'named', 'bye', 14
FROM ffl.player_season ps JOIN ffl.club_season cs ON ps.club_season_id = cs.id JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.club_match cm ON cm.club_season_id = cs.id JOIN ffl.player p ON ps.player_id = p.id JOIN afl.player ap ON p.afl_player_id = ap.id JOIN ffl.match fm ON cm.match_id = fm.id JOIN ffl.round r ON fm.round_id = r.id
WHERE c.name = 'The Howling Cows' AND ap.name = 'Hugh McCluggage' AND r.name = 'Round 5';

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
