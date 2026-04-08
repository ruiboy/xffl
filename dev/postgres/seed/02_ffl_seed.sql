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
INSERT INTO ffl.league (name) VALUES ('FFL')
ON CONFLICT (name) DO NOTHING;

-- Clubs (unique on name)
INSERT INTO ffl.club (name) VALUES
    ('Ruiboys'),
    ('The Howling Cows'),
    ('Slashers'),
    ('Cheetahs')
ON CONFLICT (name) DO NOTHING;

-- Season
INSERT INTO ffl.season (name, league_id) VALUES
    ('FFL 2026', (SELECT id FROM ffl.league WHERE name = 'FFL'));

-- Club seasons
INSERT INTO ffl.club_season (club_id, season_id) VALUES
    ((SELECT id FROM ffl.club WHERE name = 'Ruiboys'),        (SELECT id FROM ffl.season WHERE name = 'FFL 2026')),
    ((SELECT id FROM ffl.club WHERE name = 'The Howling Cows'), (SELECT id FROM ffl.season WHERE name = 'FFL 2026')),
    ((SELECT id FROM ffl.club WHERE name = 'Slashers'),       (SELECT id FROM ffl.season WHERE name = 'FFL 2026')),
    ((SELECT id FROM ffl.club WHERE name = 'Cheetahs'),       (SELECT id FROM ffl.season WHERE name = 'FFL 2026'));

-- Players: Ruiboys use Adelaide players, The Howling Cows use Brisbane players
INSERT INTO ffl.player (afl_player_id, drv_name)
SELECT ap.id, ap.name FROM afl.player ap WHERE ap.name IN ('Jordan Dawson', 'Wayne Milera');

INSERT INTO ffl.player (afl_player_id, drv_name)
SELECT ap.id, ap.name FROM afl.player ap WHERE ap.name IN (
    'Henry Smith', 'Hugh McCluggage'
);

-- Round 1 (afl_round_id references the AFL round for the same week)
INSERT INTO ffl.round (name, season_id, afl_round_id) VALUES
    ('1', (SELECT id FROM ffl.season WHERE name = 'FFL 2026'),
     (SELECT r.id FROM afl.round r JOIN afl.season s ON r.season_id = s.id WHERE r.name = 'Round 1' AND s.name = 'AFL 2026'));

-- Player seasons — Ruiboys (afl_player_season_id links to the AFL player's season entry)
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT p.id, cs.id, r.id,
       (SELECT aps.id FROM afl.player_season aps WHERE aps.player_id = ap.id LIMIT 1)
FROM ffl.player p
JOIN afl.player ap ON p.afl_player_id = ap.id
JOIN ffl.club_season cs ON cs.club_id = (SELECT id FROM ffl.club WHERE name = 'Ruiboys')
JOIN ffl.round r ON r.season_id = cs.season_id
WHERE r.name = '1' AND ap.name IN ('Jordan Dawson', 'Wayne Milera');

-- Player seasons — The Howling Cows (Henry Smith + Hugh McCluggage assigned to positions;
-- remaining 6 are squad-only with no player_match, available in the team builder)
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT p.id, cs.id, r.id,
       (SELECT aps.id FROM afl.player_season aps WHERE aps.player_id = ap.id LIMIT 1)
FROM ffl.player p
JOIN afl.player ap ON p.afl_player_id = ap.id
JOIN ffl.club_season cs ON cs.club_id = (SELECT id FROM ffl.club WHERE name = 'The Howling Cows')
JOIN ffl.round r ON r.season_id = cs.season_id
WHERE r.name = '1' AND ap.name IN (
    'Henry Smith', 'Hugh McCluggage'
);

-- Rounds 2–23 fixture (DO block for compactness)
DO $$
DECLARE
    v_season_id  INTEGER;
    v_match_id   INTEGER;
    v_home_cm_id INTEGER;
    v_away_cm_id INTEGER;
    r            RECORD;
BEGIN
    SELECT id INTO v_season_id FROM ffl.season WHERE name = 'FFL 2026';

    -- Insert rounds 2–23 in order (19 = SUPERBYE, 23 = no matches yet)
    -- afl_round_id links each FFL round to its corresponding AFL round
    INSERT INTO ffl.round (name, season_id, afl_round_id)
    SELECT n::text, v_season_id,
           (SELECT r.id FROM afl.round r JOIN afl.season s ON r.season_id = s.id
            WHERE r.name = 'Round ' || n::text AND s.name = 'AFL 2026')
    FROM generate_series(2, 23) AS n;

    -- Insert all matches (rounds 19 and 23 have none)
    FOR r IN
        SELECT * FROM (VALUES
            ('2',  '2026-03-19 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'Cheetahs'),
            ('2',  '2026-03-19 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Slashers'),
            ('3',  '2026-03-26 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Cheetahs'),
            ('3',  '2026-03-26 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'The Howling Cows'),
            ('4',  '2026-04-02 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'Slashers'),
            ('4',  '2026-04-02 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Cheetahs'),
            ('5',  '2026-04-09 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'The Howling Cows'),
            ('5',  '2026-04-09 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Ruiboys'),
            ('6',  '2026-04-16 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Ruiboys'),
            ('6',  '2026-04-16 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Slashers'),
            ('7',  '2026-04-23 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'The Howling Cows'),
            ('7',  '2026-04-23 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Ruiboys'),
            ('8',  '2026-04-30 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'Cheetahs'),
            ('8',  '2026-04-30 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Slashers'),
            ('9',  '2026-05-07 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Cheetahs'),
            ('9',  '2026-05-07 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'The Howling Cows'),
            ('10', '2026-05-14 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'Slashers'),
            ('10', '2026-05-14 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Cheetahs'),
            ('11', '2026-05-21 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'The Howling Cows'),
            ('11', '2026-05-21 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Ruiboys'),
            ('12', '2026-05-28 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Ruiboys'),
            ('12', '2026-05-28 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Slashers'),
            ('13', '2026-06-04 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'The Howling Cows'),
            ('13', '2026-06-04 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Ruiboys'),
            ('14', '2026-06-11 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'Cheetahs'),
            ('14', '2026-06-11 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Slashers'),
            ('15', '2026-06-18 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Cheetahs'),
            ('15', '2026-06-18 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'The Howling Cows'),
            ('16', '2026-06-25 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'Slashers'),
            ('16', '2026-06-25 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Cheetahs'),
            ('17', '2026-07-02 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'The Howling Cows'),
            ('17', '2026-07-02 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Ruiboys'),
            ('18', '2026-07-09 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Cheetahs'),
            ('18', '2026-07-09 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'The Howling Cows'),
            ('20', '2026-07-23 00:00:00+00'::timestamptz, 'Rui Dome',      'Ruiboys',          'Slashers'),
            ('20', '2026-07-23 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Cheetahs'),
            ('21', '2026-07-30 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'The Howling Cows'),
            ('21', '2026-07-30 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Ruiboys'),
            ('22', '2026-08-06 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Ruiboys'),
            ('22', '2026-08-06 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Slashers')
        ) AS t(round_name, start_dt, venue, home_club, away_club)
    LOOP
        INSERT INTO ffl.match (round_id, match_style, venue, start_dt)
        VALUES (
            (SELECT id FROM ffl.round WHERE name = r.round_name AND season_id = v_season_id),
            'versus', r.venue, r.start_dt
        )
        RETURNING id INTO v_match_id;

        INSERT INTO ffl.club_match (match_id, club_season_id)
        VALUES (v_match_id, (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = r.home_club AND cs.season_id = v_season_id))
        RETURNING id INTO v_home_cm_id;

        INSERT INTO ffl.club_match (match_id, club_season_id)
        VALUES (v_match_id, (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = r.away_club AND cs.season_id = v_season_id))
        RETURNING id INTO v_away_cm_id;

        UPDATE ffl.match SET home_club_match_id = v_home_cm_id, away_club_match_id = v_away_cm_id WHERE id = v_match_id;
    END LOOP;
END $$;

COMMIT;
