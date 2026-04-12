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

-- Club seasons for all 18 clubs
INSERT INTO afl.club_season (club_id, season_id)
SELECT c.id, s.id
FROM afl.club c, afl.season s
JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2026'
ON CONFLICT (club_id, season_id) DO NOTHING;

-- Players: 2 per test team
INSERT INTO afl.player (name) VALUES
('Jordan Dawson'),
('Wayne Milera'),
('Henry Smith'),
('Hugh McCluggage')
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

-- All 25 rounds: Opening Round + Rounds 1–24
INSERT INTO afl.round (season_id, name)
SELECT s.id, r.name
FROM (VALUES
    ('Opening Round'), ('Round 1'),  ('Round 2'),  ('Round 3'),  ('Round 4'),
    ('Round 5'),       ('Round 6'),  ('Round 7'),  ('Round 8'),  ('Round 9'),
    ('Round 10'),      ('Round 11'), ('Round 12'), ('Round 13'), ('Round 14'),
    ('Round 15'),      ('Round 16'), ('Round 17'), ('Round 18'), ('Round 19'),
    ('Round 20'),      ('Round 21'), ('Round 22'), ('Round 23'), ('Round 24')
) AS r(name),
afl.season s JOIN afl.league l ON s.league_id = l.id
WHERE l.name = 'AFL' AND s.name = 'AFL 2026';

-- Full 2026 fixture — all matches with home/away club_match records
DO $$
DECLARE
    v_season_id   INTEGER;
    v_match_id    INTEGER;
    v_home_cm_id  INTEGER;
    v_away_cm_id  INTEGER;
    v_match_count INTEGER := 0;
    v_cm_count    INTEGER := 0;
    rec           RECORD;
BEGIN
    SELECT s.id INTO v_season_id
    FROM afl.season s JOIN afl.league l ON s.league_id = l.id
    WHERE l.name = 'AFL' AND s.name = 'AFL 2026';

    FOR rec IN
        SELECT * FROM (VALUES
            -- Opening Round
            ('Opening Round', '2026-03-05 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Carlton Blues'),
            ('Opening Round', '2026-03-06 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'Geelong Cats'),
            ('Opening Round', '2026-03-07 00:00:00+00'::timestamptz, 'ENGIE Stadium',                   'Greater Western Sydney Giants',   'Hawthorn Hawks'),
            ('Opening Round', '2026-03-07 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Western Bulldogs'),
            ('Opening Round', '2026-03-08 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'St Kilda Saints',                 'Collingwood Magpies'),
            -- Round 1
            ('Round 1',  '2026-03-12 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Carlton Blues',                   'Richmond Tigers'),
            ('Round 1',  '2026-03-13 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Essendon Bombers',                'Hawthorn Hawks'),
            ('Round 1',  '2026-03-14 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Greater Western Sydney Giants'),
            ('Round 1',  '2026-03-14 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'Fremantle Dockers'),
            ('Round 1',  '2026-03-14 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Brisbane Lions'),
            ('Round 1',  '2026-03-14 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'Adelaide Crows'),
            ('Round 1',  '2026-03-15 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'Port Adelaide Power'),
            ('Round 1',  '2026-03-15 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'St Kilda Saints'),
            ('Round 1',  '2026-03-15 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'West Coast Eagles'),
            -- Round 2  (byes: Brisbane Lions, Carlton Blues, Collingwood Magpies, Geelong Cats)
            ('Round 2',  '2026-03-19 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Hawthorn Hawks',                  'Sydney Swans'),
            ('Round 2',  '2026-03-20 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Western Bulldogs'),
            ('Round 2',  '2026-03-21 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'Gold Coast Suns'),
            ('Round 2',  '2026-03-21 00:00:00+00'::timestamptz, 'ENGIE Stadium',                   'Greater Western Sydney Giants',   'St Kilda Saints'),
            ('Round 2',  '2026-03-21 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Melbourne Demons'),
            ('Round 2',  '2026-03-22 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Essendon Bombers'),
            ('Round 2',  '2026-03-22 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'North Melbourne Kangaroos'),
            -- Round 3  (byes: Gold Coast Suns, Hawthorn Hawks, Sydney Swans, Western Bulldogs)
            ('Round 3',  '2026-03-26 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'Adelaide Crows'),
            ('Round 3',  '2026-03-27 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Collingwood Magpies',             'Greater Western Sydney Giants'),
            ('Round 3',  '2026-03-28 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Brisbane Lions'),
            ('Round 3',  '2026-03-28 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Richmond Tigers'),
            ('Round 3',  '2026-03-28 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Essendon Bombers',                'North Melbourne Kangaroos'),
            ('Round 3',  '2026-03-29 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'West Coast Eagles'),
            ('Round 3',  '2026-03-29 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Carlton Blues',                   'Melbourne Demons'),
            -- Round 4  (byes: Greater Western Sydney Giants, St Kilda Saints)
            ('Round 4',  '2026-04-02 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Collingwood Magpies'),
            ('Round 4',  '2026-04-03 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'Carlton Blues'),
            ('Round 4',  '2026-04-03 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Fremantle Dockers'),
            ('Round 4',  '2026-04-04 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'Port Adelaide Power'),
            ('Round 4',  '2026-04-04 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Sydney Swans'),
            ('Round 4',  '2026-04-05 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'Gold Coast Suns'),
            ('Round 4',  '2026-04-05 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Essendon Bombers'),
            ('Round 4',  '2026-04-06 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Hawthorn Hawks',                  'Geelong Cats'),
            -- Round 5
            ('Round 5',  '2026-04-09 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Carlton Blues'),
            ('Round 5',  '2026-04-10 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Collingwood Magpies',             'Fremantle Dockers'),
            ('Round 5',  '2026-04-11 00:00:00+00'::timestamptz, 'Barossa Park',                    'North Melbourne Kangaroos',       'Brisbane Lions'),
            ('Round 5',  '2026-04-11 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Essendon Bombers',                'Melbourne Demons'),
            ('Round 5',  '2026-04-11 00:00:00+00'::timestamptz, 'Norwood Oval',                    'Sydney Swans',                    'Gold Coast Suns'),
            ('Round 5',  '2026-04-11 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Hawthorn Hawks',                  'Western Bulldogs'),
            ('Round 5',  '2026-04-12 00:00:00+00'::timestamptz, 'Norwood Oval',                    'Geelong Cats',                    'West Coast Eagles'),
            ('Round 5',  '2026-04-12 00:00:00+00'::timestamptz, 'Barossa Park',                    'Greater Western Sydney Giants',   'Richmond Tigers'),
            ('Round 5',  '2026-04-12 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'St Kilda Saints'),
            -- Round 6
            ('Round 6',  '2026-04-16 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Carlton Blues',                   'Collingwood Magpies'),
            ('Round 6',  '2026-04-17 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'Western Bulldogs'),
            ('Round 6',  '2026-04-17 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Greater Western Sydney Giants'),
            ('Round 6',  '2026-04-18 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'Essendon Bombers'),
            ('Round 6',  '2026-04-18 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Hawthorn Hawks',                  'Port Adelaide Power'),
            ('Round 6',  '2026-04-18 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'St Kilda Saints'),
            ('Round 6',  '2026-04-19 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'Richmond Tigers'),
            ('Round 6',  '2026-04-19 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'Brisbane Lions'),
            ('Round 6',  '2026-04-19 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Fremantle Dockers'),
            -- Round 7
            ('Round 7',  '2026-04-23 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Sydney Swans'),
            ('Round 7',  '2026-04-24 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'Melbourne Demons'),
            ('Round 7',  '2026-04-25 00:00:00+00'::timestamptz, 'University of Tasmania Stadium',  'Hawthorn Hawks',                  'Gold Coast Suns'),
            ('Round 7',  '2026-04-25 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Essendon Bombers',                'Collingwood Magpies'),
            ('Round 7',  '2026-04-25 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Geelong Cats'),
            ('Round 7',  '2026-04-25 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Carlton Blues'),
            ('Round 7',  '2026-04-26 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'West Coast Eagles'),
            ('Round 7',  '2026-04-26 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Adelaide Crows'),
            ('Round 7',  '2026-04-26 00:00:00+00'::timestamptz, 'Corroboree Group Oval Manuk',     'Greater Western Sydney Giants',   'North Melbourne Kangaroos'),
            -- Round 8
            ('Round 8',  '2026-04-30 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'Hawthorn Hawks'),
            ('Round 8',  '2026-05-01 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Fremantle Dockers'),
            ('Round 8',  '2026-05-01 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Port Adelaide Power'),
            ('Round 8',  '2026-05-02 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Essendon Bombers',                'Brisbane Lions'),
            ('Round 8',  '2026-05-02 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Richmond Tigers'),
            ('Round 8',  '2026-05-02 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'North Melbourne Kangaroos'),
            ('Round 8',  '2026-05-02 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Carlton Blues',                   'St Kilda Saints'),
            ('Round 8',  '2026-05-03 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Melbourne Demons'),
            ('Round 8',  '2026-05-03 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'Greater Western Sydney Giants'),
            -- Round 9
            ('Round 9',  '2026-05-07 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Hawthorn Hawks'),
            ('Round 9',  '2026-05-08 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Carlton Blues'),
            ('Round 9',  '2026-05-08 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Western Bulldogs'),
            ('Round 9',  '2026-05-09 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'Sydney Swans'),
            ('Round 9',  '2026-05-09 00:00:00+00'::timestamptz, 'ENGIE Stadium',                   'Greater Western Sydney Giants',   'Essendon Bombers'),
            ('Round 9',  '2026-05-09 00:00:00+00'::timestamptz, 'TIO Stadium',                     'Gold Coast Suns',                 'St Kilda Saints'),
            ('Round 9',  '2026-05-09 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Geelong Cats',                    'Collingwood Magpies'),
            ('Round 9',  '2026-05-10 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Melbourne Demons',                'West Coast Eagles'),
            ('Round 9',  '2026-05-10 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'Adelaide Crows'),
            -- Round 10
            ('Round 10', '2026-05-14 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Geelong Cats'),
            ('Round 10', '2026-05-15 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Collingwood Magpies'),
            ('Round 10', '2026-05-15 00:00:00+00'::timestamptz, 'TIO Stadium',                     'Gold Coast Suns',                 'Port Adelaide Power'),
            ('Round 10', '2026-05-16 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'North Melbourne Kangaroos'),
            ('Round 10', '2026-05-16 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'Hawthorn Hawks'),
            ('Round 10', '2026-05-16 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Carlton Blues',                   'Western Bulldogs'),
            ('Round 10', '2026-05-17 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Essendon Bombers',                'Fremantle Dockers'),
            ('Round 10', '2026-05-17 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Richmond Tigers'),
            ('Round 10', '2026-05-17 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Greater Western Sydney Giants'),
            -- Round 11
            ('Round 11', '2026-05-21 00:00:00+00'::timestamptz, 'University of Tasmania Stadium',  'Hawthorn Hawks',                  'Adelaide Crows'),
            ('Round 11', '2026-05-22 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'Essendon Bombers'),
            ('Round 11', '2026-05-22 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'St Kilda Saints'),
            ('Round 11', '2026-05-23 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'Gold Coast Suns'),
            ('Round 11', '2026-05-23 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'Sydney Swans'),
            ('Round 11', '2026-05-23 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'West Coast Eagles'),
            ('Round 11', '2026-05-23 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Carlton Blues'),
            ('Round 11', '2026-05-24 00:00:00+00'::timestamptz, 'ENGIE Stadium',                   'Greater Western Sydney Giants',   'Brisbane Lions'),
            ('Round 11', '2026-05-24 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Melbourne Demons'),
            -- Round 12  (byes: Adelaide Crows, Gold Coast Suns, North Melbourne Kangaroos, Port Adelaide Power)
            ('Round 12', '2026-05-28 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Hawthorn Hawks'),
            ('Round 12', '2026-05-29 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Carlton Blues',                   'Geelong Cats'),
            ('Round 12', '2026-05-30 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Richmond Tigers'),
            ('Round 12', '2026-05-30 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Fremantle Dockers'),
            ('Round 12', '2026-05-30 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Collingwood Magpies'),
            ('Round 12', '2026-05-31 00:00:00+00'::timestamptz, 'TIO Traeger Park',                'Melbourne Demons',                'Greater Western Sydney Giants'),
            ('Round 12', '2026-05-31 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Essendon Bombers'),
            -- Round 13  (byes: Greater Western Sydney Giants, Richmond Tigers)
            ('Round 13', '2026-06-04 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Geelong Cats'),
            ('Round 13', '2026-06-05 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Hawthorn Hawks',                  'Western Bulldogs'),
            ('Round 13', '2026-06-06 00:00:00+00'::timestamptz, 'Hands Oval',                      'North Melbourne Kangaroos',       'Fremantle Dockers'),
            ('Round 13', '2026-06-06 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'Brisbane Lions'),
            ('Round 13', '2026-06-06 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Port Adelaide Power'),
            ('Round 13', '2026-06-07 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'St Kilda Saints'),
            ('Round 13', '2026-06-07 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Essendon Bombers',                'Carlton Blues'),
            ('Round 13', '2026-06-08 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'Melbourne Demons'),
            -- Round 14  (byes: Carlton Blues, Collingwood Magpies, Fremantle Dockers, Hawthorn Hawks)
            ('Round 14', '2026-06-11 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Adelaide Crows'),
            ('Round 14', '2026-06-12 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'Gold Coast Suns'),
            ('Round 14', '2026-06-13 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'Essendon Bombers'),
            ('Round 14', '2026-06-13 00:00:00+00'::timestamptz, 'Optus Stadium',                   'North Melbourne Kangaroos',       'West Coast Eagles'),
            ('Round 14', '2026-06-13 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Sydney Swans'),
            ('Round 14', '2026-06-14 00:00:00+00'::timestamptz, 'Ninja Stadium',                   'Richmond Tigers',                 'Brisbane Lions'),
            ('Round 14', '2026-06-14 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Greater Western Sydney Giants'),
            -- Round 15  (byes: Brisbane Lions, Essendon Bombers, Sydney Swans, West Coast Eagles)
            ('Round 15', '2026-06-18 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Geelong Cats'),
            ('Round 15', '2026-06-19 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'Hawthorn Hawks'),
            ('Round 15', '2026-06-20 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Melbourne Demons'),
            ('Round 15', '2026-06-20 00:00:00+00'::timestamptz, 'ENGIE Stadium',                   'Greater Western Sydney Giants',   'Carlton Blues'),
            ('Round 15', '2026-06-20 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'Port Adelaide Power'),
            ('Round 15', '2026-06-21 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'North Melbourne Kangaroos'),
            ('Round 15', '2026-06-21 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Western Bulldogs'),
            -- Round 16  (byes: Geelong Cats, Melbourne Demons, St Kilda Saints, Western Bulldogs) — date range Thu Jun 25–Sun Jun 28
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Sydney Swans'),
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Carlton Blues',                   'West Coast Eagles'),
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'Richmond Tigers'),
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Gold Coast Suns'),
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Hawthorn Hawks',                  'Greater Western Sydney Giants'),
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'Essendon Bombers'),
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Adelaide Crows'),
            -- Round 17  — date range Thu Jul 2–Sun Jul 5
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Essendon Bombers',                'St Kilda Saints'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'Brisbane Lions'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'Collingwood Magpies'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'Corroboree Group Oval Manuk',     'Greater Western Sydney Giants',   'Fremantle Dockers'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'University of Tasmania Stadium',  'Hawthorn Hawks',                  'Melbourne Demons'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'North Melbourne Kangaroos'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'Carlton Blues'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Western Bulldogs'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Adelaide Crows'),
            -- Round 18  — date range Thu Jul 9–Sun Jul 12
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Gold Coast Suns'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Essendon Bombers'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Carlton Blues',                   'Hawthorn Hawks'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Collingwood Magpies',             'North Melbourne Kangaroos'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Sydney Swans'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'ENGIE Stadium',                   'Greater Western Sydney Giants',   'Geelong Cats'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'Richmond Tigers'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Port Adelaide Power'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'West Coast Eagles'),
            -- Round 19  — date range Thu Jul 16–Sun Jul 19
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'Carlton Blues'),
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Essendon Bombers',                'Greater Western Sydney Giants'),
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'St Kilda Saints'),
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'Western Bulldogs'),
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'Melbourne Demons'),
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Fremantle Dockers'),
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'Hawthorn Hawks'),
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Adelaide Crows'),
            ('Round 19', '2026-07-16 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Brisbane Lions'),
            -- Round 20  — date range Thu Jul 23–Sun Jul 26
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Collingwood Magpies'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Port Adelaide Power'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Carlton Blues',                   'Gold Coast Suns'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'West Coast Eagles'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'ENGIE Stadium',                   'Greater Western Sydney Giants',   'Sydney Swans'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Hawthorn Hawks',                  'Essendon Bombers'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'St Kilda Saints'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'Geelong Cats'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Richmond Tigers'),
            -- Round 21  — date range Thu Jul 30–Sun Aug 2
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Carlton Blues',                   'Brisbane Lions'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'Geelong Cats'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Essendon Bombers',                'Adelaide Crows'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Western Bulldogs'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'People First Stadium',            'Gold Coast Suns',                 'Melbourne Demons'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'University of Tasmania Stadium',  'Hawthorn Hawks',                  'North Melbourne Kangaroos'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Greater Western Sydney Giants'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'West Coast Eagles'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Sydney Swans'),
            -- Round 22  — date range Thu Aug 6–Sun Aug 9
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Richmond Tigers'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Hawthorn Hawks'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'Essendon Bombers'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Corroboree Group Oval Manuk',     'Greater Western Sydney Giants',   'Gold Coast Suns'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'Fremantle Dockers'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Carlton Blues'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'Port Adelaide Power'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Collingwood Magpies'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'North Melbourne Kangaroos'),
            -- Round 23  — date range Fri Aug 14–Sun Aug 16
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'The Gabba',                       'Brisbane Lions',                  'Gold Coast Suns'),
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Essendon Bombers',                'Sydney Swans'),
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'Optus Stadium',                   'Fremantle Dockers',               'Adelaide Crows'),
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'ENGIE Stadium',                   'Greater Western Sydney Giants',   'West Coast Eagles'),
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Hawthorn Hawks',                  'Collingwood Magpies'),
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'North Melbourne Kangaroos',       'Geelong Cats'),
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Port Adelaide Power',             'Melbourne Demons'),
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Richmond Tigers',                 'St Kilda Saints'),
            ('Round 23', '2026-08-14 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Western Bulldogs',                'Carlton Blues'),
            -- Round 24  — date range Thu Aug 21–Sun Aug 23
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'Adelaide Oval',                   'Adelaide Crows',                  'Greater Western Sydney Giants'),
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Collingwood Magpies',             'Brisbane Lions'),
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Carlton Blues',                   'Fremantle Dockers'),
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'Essendon Bombers',                'Port Adelaide Power'),
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'GMHBA Stadium',                   'Geelong Cats',                    'Richmond Tigers'),
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'Melbourne Cricket Ground',        'Melbourne Demons',                'Western Bulldogs'),
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'Marvel Stadium',                  'St Kilda Saints',                 'Gold Coast Suns'),
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'Sydney Cricket Ground',           'Sydney Swans',                    'North Melbourne Kangaroos'),
            ('Round 24', '2026-08-21 00:00:00+00'::timestamptz, 'Optus Stadium',                   'West Coast Eagles',               'Hawthorn Hawks')
        ) AS t(round_name, start_dt, venue, home_club, away_club)
    LOOP
        INSERT INTO afl.match (round_id, venue, start_dt, drv_result)
        VALUES (
            (SELECT r.id FROM afl.round r WHERE r.name = rec.round_name AND r.season_id = v_season_id),
            rec.venue, rec.start_dt, 'no_result'
        )
        RETURNING id INTO v_match_id;
        v_match_count := v_match_count + 1;

        INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
        VALUES (v_match_id,
            (SELECT cs.id FROM afl.club_season cs JOIN afl.club c ON cs.club_id = c.id
             WHERE c.name = rec.home_club AND cs.season_id = v_season_id),
            0, 0, 0)
        RETURNING id INTO v_home_cm_id;
        v_cm_count := v_cm_count + 1;

        INSERT INTO afl.club_match (match_id, club_season_id, drv_score, drv_premiership_points, rushed_behinds)
        VALUES (v_match_id,
            (SELECT cs.id FROM afl.club_season cs JOIN afl.club c ON cs.club_id = c.id
             WHERE c.name = rec.away_club AND cs.season_id = v_season_id),
            0, 0, 0)
        RETURNING id INTO v_away_cm_id;
        v_cm_count := v_cm_count + 1;

        UPDATE afl.match
        SET home_club_match_id = v_home_cm_id, away_club_match_id = v_away_cm_id
        WHERE id = v_match_id;
    END LOOP;

    RAISE NOTICE 'matches inserted: %', v_match_count;
    RAISE NOTICE 'club_matches inserted: %', v_cm_count;
END $$;

-- Test player_match data for FFL score computation testing
-- Round 1: Collingwood Magpies vs Adelaide Crows @ MCG (Adelaide away)
INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 18, 12, 6, 0, 4, 2, 1
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
JOIN afl.match m ON cm.match_id = m.id
JOIN afl.round r ON m.round_id = r.id
WHERE p.name = 'Jordan Dawson' AND c.name = 'Adelaide Crows' AND r.name = 'Round 1'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 22, 15, 4, 0, 6, 0, 2
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
JOIN afl.match m ON cm.match_id = m.id
JOIN afl.round r ON m.round_id = r.id
WHERE p.name = 'Wayne Milera' AND c.name = 'Adelaide Crows' AND r.name = 'Round 1'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

-- Round 1: Sydney Swans vs Brisbane Lions @ SCG (Brisbane away)
INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 20, 16, 5, 0, 8, 3, 2
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
JOIN afl.match m ON cm.match_id = m.id
JOIN afl.round r ON m.round_id = r.id
WHERE p.name = 'Henry Smith' AND c.name = 'Brisbane Lions' AND r.name = 'Round 1'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 16, 10, 7, 0, 5, 1, 3
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
JOIN afl.match m ON cm.match_id = m.id
JOIN afl.round r ON m.round_id = r.id
WHERE p.name = 'Hugh McCluggage' AND c.name = 'Brisbane Lions' AND r.name = 'Round 1'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

-- Round 2: Adelaide Crows vs Western Bulldogs @ Adelaide Oval
INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 15, 10, 8, 0, 3, 1, 2
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
JOIN afl.match m ON cm.match_id = m.id
JOIN afl.round r ON m.round_id = r.id
WHERE p.name = 'Jordan Dawson' AND c.name = 'Adelaide Crows' AND r.name = 'Round 2'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

INSERT INTO afl.player_match (player_season_id, club_match_id, status, kicks, handballs, marks, hitouts, tackles, goals, behinds)
SELECT ps.id, cm.id, 'played', 25, 11, 5, 0, 8, 1, 0
FROM afl.player_season ps
JOIN afl.player p ON ps.player_id = p.id
JOIN afl.club_season cs ON ps.club_season_id = cs.id
JOIN afl.club c ON cs.club_id = c.id
JOIN afl.club_match cm ON cm.club_season_id = cs.id
JOIN afl.match m ON cm.match_id = m.id
JOIN afl.round r ON m.round_id = r.id
WHERE p.name = 'Wayne Milera' AND c.name = 'Adelaide Crows' AND r.name = 'Round 2'
ON CONFLICT (player_season_id, club_match_id) DO NOTHING;

COMMIT;
