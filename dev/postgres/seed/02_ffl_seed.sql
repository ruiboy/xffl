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
INSERT INTO ffl.season (name, league_id, afl_season_id) VALUES
    ('FFL 2026', (SELECT id FROM ffl.league WHERE name = 'FFL'),
     (SELECT id FROM afl.season WHERE name = 'AFL 2026'));

-- Club seasons
INSERT INTO ffl.club_season (club_id, season_id) VALUES
    ((SELECT id FROM ffl.club WHERE name = 'Ruiboys'),        (SELECT id FROM ffl.season WHERE name = 'FFL 2026')),
    ((SELECT id FROM ffl.club WHERE name = 'The Howling Cows'), (SELECT id FROM ffl.season WHERE name = 'FFL 2026')),
    ((SELECT id FROM ffl.club WHERE name = 'Slashers'),       (SELECT id FROM ffl.season WHERE name = 'FFL 2026')),
    ((SELECT id FROM ffl.club WHERE name = 'Cheetahs'),       (SELECT id FROM ffl.season WHERE name = 'FFL 2026'));

-- Round 1 (afl_round_id references the AFL round for the same week)
INSERT INTO ffl.round (name, season_id, afl_round_id) VALUES
    ('Round 1', (SELECT id FROM ffl.season WHERE name = 'FFL 2026'),
     (SELECT r.id FROM afl.round r JOIN afl.season s ON r.season_id = s.id WHERE r.name = 'Round 1' AND s.name = 'AFL 2026'));

-- Rounds 2–23 fixture (DO block for compactness)
DO $$
DECLARE
    v_season_id  INTEGER;
    v_match_id   INTEGER;
    v_home_cm_id  INTEGER;
    v_away_cm_id  INTEGER;
    v_match_count INTEGER := 0;
    v_cm_count    INTEGER := 0;
    v_row_count   INTEGER;
    rec           RECORD;
BEGIN
    SELECT id INTO v_season_id FROM ffl.season WHERE name = 'FFL 2026';

    -- Insert rounds 2–23 in order (19 = SUPERBYE, 23 = no matches yet)
    -- afl_round_id links each FFL round to its corresponding AFL round
    INSERT INTO ffl.round (name, season_id, afl_round_id)
    SELECT 'Round ' || n::text, v_season_id,
           (SELECT ar.id FROM afl.round ar JOIN afl.season s ON ar.season_id = s.id
            WHERE ar.name = 'Round ' || n::text AND s.name = 'AFL 2026')
    FROM generate_series(2, 23) AS n;
    GET DIAGNOSTICS v_row_count = ROW_COUNT;
    RAISE NOTICE 'rounds inserted: %', v_row_count;

    -- Insert all matches (rounds 19 and 23 have none)
    FOR rec IN
        SELECT * FROM (VALUES
            ('Round 1',  '2026-03-13 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'The Howling Cows'),
            ('Round 1',  '2026-03-13 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Ruiboys'),
            ('Round 2',  '2026-03-19 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'Cheetahs'),
            ('Round 2',  '2026-03-19 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Slashers'),
            ('Round 3',  '2026-03-26 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Cheetahs'),
            ('Round 3',  '2026-03-26 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'The Howling Cows'),
            ('Round 4',  '2026-04-02 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'Slashers'),
            ('Round 4',  '2026-04-02 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Cheetahs'),
            ('Round 5',  '2026-04-09 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'The Howling Cows'),
            ('Round 5',  '2026-04-09 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Ruiboys'),
            ('Round 6',  '2026-04-16 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Ruiboys'),
            ('Round 6',  '2026-04-16 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Slashers'),
            ('Round 7',  '2026-04-23 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'The Howling Cows'),
            ('Round 7',  '2026-04-23 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Ruiboys'),
            ('Round 8',  '2026-04-30 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'Cheetahs'),
            ('Round 8',  '2026-04-30 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Slashers'),
            ('Round 9',  '2026-05-07 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Cheetahs'),
            ('Round 9',  '2026-05-07 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'The Howling Cows'),
            ('Round 10', '2026-05-14 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'Slashers'),
            ('Round 10', '2026-05-14 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Cheetahs'),
            ('Round 11', '2026-05-21 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'The Howling Cows'),
            ('Round 11', '2026-05-21 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Ruiboys'),
            ('Round 12', '2026-05-28 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Ruiboys'),
            ('Round 12', '2026-05-28 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Slashers'),
            ('Round 13', '2026-06-04 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'The Howling Cows'),
            ('Round 13', '2026-06-04 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Ruiboys'),
            ('Round 14', '2026-06-11 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'Cheetahs'),
            ('Round 14', '2026-06-11 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Slashers'),
            ('Round 15', '2026-06-18 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Cheetahs'),
            ('Round 15', '2026-06-18 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'The Howling Cows'),
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'Slashers'),
            ('Round 16', '2026-06-25 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Cheetahs'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'The Howling Cows'),
            ('Round 17', '2026-07-02 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Ruiboys'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'Cheetahs'),
            ('Round 18', '2026-07-09 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'The Howling Cows'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'The Rui Inn',      'Ruiboys',          'Slashers'),
            ('Round 20', '2026-07-23 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Cheetahs'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'The Slash Pit', 'Slashers',         'The Howling Cows'),
            ('Round 21', '2026-07-30 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Ruiboys'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Moo Meadow',    'The Howling Cows', 'Ruiboys'),
            ('Round 22', '2026-08-06 00:00:00+00'::timestamptz, 'Savanna Park',  'Cheetahs',         'Slashers')
        ) AS t(round_name, start_dt, venue, home_club, away_club)
    LOOP
        INSERT INTO ffl.match (round_id, match_style, venue, start_dt)
        VALUES (
            (SELECT id FROM ffl.round WHERE name = rec.round_name AND season_id = v_season_id),
            'versus', rec.venue, rec.start_dt
        )
        RETURNING id INTO v_match_id;
        v_match_count := v_match_count + 1;

        INSERT INTO ffl.club_match (match_id, club_season_id)
        VALUES (v_match_id, (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = rec.home_club AND cs.season_id = v_season_id))
        RETURNING id INTO v_home_cm_id;
        v_cm_count := v_cm_count + 1;

        INSERT INTO ffl.club_match (match_id, club_season_id)
        VALUES (v_match_id, (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id WHERE c.name = rec.away_club AND cs.season_id = v_season_id))
        RETURNING id INTO v_away_cm_id;
        v_cm_count := v_cm_count + 1;

        UPDATE ffl.match SET home_club_match_id = v_home_cm_id, away_club_match_id = v_away_cm_id WHERE id = v_match_id;
    END LOOP;
    RAISE NOTICE 'matches inserted: %', v_match_count;
    RAISE NOTICE 'club_matches inserted: %', v_cm_count;
END $$;

COMMIT;
