-- FFL player rosters for 2026
-- Runs after 03_afl_historical.sql so all AFL player records exist
BEGIN;

-- Ruiboys squad: link each player to their AFL record
INSERT INTO ffl.player (afl_player_id, drv_name)
SELECT ap.id, ap.name FROM afl.player ap
WHERE ap.name IN (
    'Jeremy Cameron', 'Sam Darcy', 'Mitch Georgiades', 'Archie Roberts',
    'Lachie Ash', 'Bailey Dale', 'Josh Daicos', 'Marcus Windhager',
    'Matthew Kennedy', 'Tim Taranto', 'Willem Drew', 'Callum Wilkie',
    'Harris Andrews', 'Tom Atkins', 'Ned Long', 'Max Gawn',
    'Luke Jackson', 'Jye Caldwell', 'Hugh McCluggage', 'Karl Amon',
    'Taylor Walker', 'Connor Idun', 'Tom Powell', 'Alex Davies',
    'George Hewett', 'Jamarra Ugle-Hagan', 'Jack Macrae', 'Reilly OBrien',
    'Cooper Lord', 'Dayne Zorko'
);

-- Ruiboys player_seasons: link to Ruiboys FFL club_season and each player's AFL 2026 season
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT
    fp.id,
    (SELECT cs.id FROM ffl.club_season cs
     JOIN ffl.club c ON cs.club_id = c.id
     JOIN ffl.season s ON cs.season_id = s.id
     WHERE c.name = 'Ruiboys' AND s.name = 'FFL 2026'),
    (SELECT r.id FROM ffl.round r
     JOIN ffl.season s ON r.season_id = s.id
     WHERE r.name = 'Round 1' AND s.name = 'FFL 2026'),
    (SELECT aps.id FROM afl.player_season aps
     JOIN afl.club_season acs ON aps.club_season_id = acs.id
     JOIN afl.club ac ON acs.club_id = ac.id
     JOIN afl.season asn ON acs.season_id = asn.id
     WHERE aps.player_id = ap.id AND asn.name = 'AFL 2026' AND ac.name = v.afl_club)
FROM (VALUES
    ('Jeremy Cameron',    'Geelong Cats'),
    ('Sam Darcy',         'Western Bulldogs'),
    ('Mitch Georgiades',  'Port Adelaide Power'),
    ('Archie Roberts',    'Essendon Bombers'),
    ('Lachie Ash',        'Greater Western Sydney Giants'),
    ('Bailey Dale',       'Western Bulldogs'),
    ('Josh Daicos',       'Collingwood Magpies'),
    ('Marcus Windhager',  'St Kilda Saints'),
    ('Matthew Kennedy',   'Western Bulldogs'),
    ('Tim Taranto',       'Richmond Tigers'),
    ('Willem Drew',       'Port Adelaide Power'),
    ('Callum Wilkie',     'St Kilda Saints'),
    ('Harris Andrews',    'Brisbane Lions'),
    ('Tom Atkins',        'Geelong Cats'),
    ('Ned Long',          'Collingwood Magpies'),
    ('Max Gawn',          'Melbourne Demons'),
    ('Luke Jackson',      'Fremantle Dockers'),
    ('Jye Caldwell',      'Essendon Bombers'),
    ('Hugh McCluggage',   'Brisbane Lions'),
    ('Karl Amon',         'Hawthorn Hawks'),
    ('Taylor Walker',     'Adelaide Crows'),
    ('Connor Idun',       'Greater Western Sydney Giants'),
    ('Tom Powell',        'North Melbourne Kangaroos'),
    ('Alex Davies',       'Gold Coast Suns'),
    ('George Hewett',     'Carlton Blues'),
    ('Jamarra Ugle-Hagan','Gold Coast Suns'),
    ('Jack Macrae',       'St Kilda Saints'),
    ('Reilly OBrien',     'Adelaide Crows'),
    ('Cooper Lord',       'Carlton Blues'),
    ('Dayne Zorko',       'Brisbane Lions')
) AS v(player_name, afl_club)
JOIN afl.player ap ON ap.name = v.player_name
JOIN ffl.player fp ON fp.afl_player_id = ap.id
ON CONFLICT (player_id, club_season_id) DO NOTHING;

COMMIT;