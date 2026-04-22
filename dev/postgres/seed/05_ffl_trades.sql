-- Mid-season FFL trades: players not in initial squads or historical data
-- Runs after 04_ffl_players.sql
BEGIN;

-- Add to afl.player (absent from historical data due to injury/omission)
INSERT INTO afl.player (name) VALUES
    ('Dan Houston'),
    ('Luke Davies-Uniacke'),
    ('Dante Visentini'),
    ('Josh Treacy')
ON CONFLICT DO NOTHING;

-- Add to ffl.player
INSERT INTO ffl.player (afl_player_id, drv_name)
SELECT ap.id, ap.name FROM afl.player ap
WHERE ap.name IN ('Dan Houston', 'Luke Davies-Uniacke', 'Dante Visentini', 'Josh Treacy')
AND NOT EXISTS (SELECT 1 FROM ffl.player fp WHERE fp.afl_player_id = ap.id);

-- Dan Houston → Cheetahs from Round 2
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT fp.id,
    (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.season s ON cs.season_id = s.id WHERE c.name = 'Cheetahs' AND s.name = 'FFL 2026'),
    (SELECT r.id FROM ffl.round r JOIN ffl.season s ON r.season_id = s.id WHERE r.name = 'Round 2' AND s.name = 'FFL 2026'),
    NULL
FROM ffl.player fp JOIN afl.player ap ON fp.afl_player_id = ap.id
WHERE ap.name = 'Dan Houston'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

-- Luke Davies-Uniacke → The Howling Cows from Round 3
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT fp.id,
    (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.season s ON cs.season_id = s.id WHERE c.name = 'The Howling Cows' AND s.name = 'FFL 2026'),
    (SELECT r.id FROM ffl.round r JOIN ffl.season s ON r.season_id = s.id WHERE r.name = 'Round 3' AND s.name = 'FFL 2026'),
    NULL
FROM ffl.player fp JOIN afl.player ap ON fp.afl_player_id = ap.id
WHERE ap.name = 'Luke Davies-Uniacke'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

-- Dante Visentini → The Howling Cows from Round 3
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT fp.id,
    (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.season s ON cs.season_id = s.id WHERE c.name = 'The Howling Cows' AND s.name = 'FFL 2026'),
    (SELECT r.id FROM ffl.round r JOIN ffl.season s ON r.season_id = s.id WHERE r.name = 'Round 3' AND s.name = 'FFL 2026'),
    NULL
FROM ffl.player fp JOIN afl.player ap ON fp.afl_player_id = ap.id
WHERE ap.name = 'Dante Visentini'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

-- Josh Treacy → Slashers from Round 4
INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)
SELECT fp.id,
    (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.season s ON cs.season_id = s.id WHERE c.name = 'Slashers' AND s.name = 'FFL 2026'),
    (SELECT r.id FROM ffl.round r JOIN ffl.season s ON r.season_id = s.id WHERE r.name = 'Round 4' AND s.name = 'FFL 2026'),
    NULL
FROM ffl.player fp JOIN afl.player ap ON fp.afl_player_id = ap.id
WHERE ap.name = 'Josh Treacy'
ON CONFLICT (player_id, club_season_id) DO NOTHING;

COMMIT;
