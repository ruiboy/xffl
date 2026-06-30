-- Compute all derived and status fields after base seed data is loaded.
-- Runs last; depends on all prior seed files.
BEGIN;

-- ============================================================
-- AFL: Compute club_match scores from player stats
-- ============================================================

UPDATE afl.club_match cm
SET drv_score = (
    SELECT COALESCE(SUM(pm.goals * 6 + pm.behinds), 0)
    FROM afl.player_match pm
    WHERE pm.club_match_id = cm.id
) + COALESCE(cm.rushed_behinds, 0)
WHERE cm.id IN (
    SELECT cm2.id FROM afl.club_match cm2
    JOIN afl.match m ON cm2.match_id = m.id
    JOIN afl.round r ON m.round_id = r.id
    JOIN afl.season s ON r.season_id = s.id
    WHERE s.name = 'AFL 2026'
    AND r.name IN ('Opening Round', 'Round 1')
);

-- ============================================================
-- AFL: Set data_status = 'final' and compute drv_result
-- ============================================================

UPDATE afl.match m
SET
    data_status = 'final',
    drv_result = CASE
        WHEN home_cm.drv_score > away_cm.drv_score THEN 'home_win'
        WHEN away_cm.drv_score > home_cm.drv_score THEN 'away_win'
        ELSE 'draw'
    END
FROM afl.club_match home_cm,
     afl.club_match away_cm
WHERE home_cm.match_id = m.id AND home_cm.side = 'home'
AND   away_cm.match_id = m.id AND away_cm.side = 'away'
AND m.id IN (
    SELECT m2.id FROM afl.match m2
    JOIN afl.round r ON m2.round_id = r.id
    JOIN afl.season s ON r.season_id = s.id
    WHERE s.name = 'AFL 2026'
    AND r.name IN ('Opening Round', 'Round 1')
);

-- ============================================================
-- AFL: Set club_match premiership points
-- ============================================================

UPDATE afl.club_match cm
SET drv_premiership_points = CASE
    WHEN m.drv_result = 'home_win' AND cm.side = 'home' THEN 4
    WHEN m.drv_result = 'away_win' AND cm.side = 'away' THEN 4
    WHEN m.drv_result = 'draw' THEN 2
    ELSE 0
END
FROM afl.match m
JOIN afl.round r ON m.round_id = r.id
JOIN afl.season s ON r.season_id = s.id
WHERE cm.match_id = m.id
AND m.data_status = 'final'
AND s.name = 'AFL 2026';

-- ============================================================
-- AFL: Update club_season ladder stats
-- ============================================================

UPDATE afl.club_season cs
SET
    drv_played             = stats.played,
    drv_won                = stats.won,
    drv_lost               = stats.lost,
    drv_drawn              = stats.drawn,
    drv_for                = stats.for_pts,
    drv_against            = stats.against_pts,
    drv_premiership_points = stats.pp
FROM (
    SELECT
        cm.club_season_id,
        COUNT(*)                                                       AS played,
        COUNT(*) FILTER (WHERE cm.drv_premiership_points = 4)         AS won,
        COUNT(*) FILTER (WHERE cm.drv_premiership_points = 0)         AS lost,
        COUNT(*) FILTER (WHERE cm.drv_premiership_points = 2)         AS drawn,
        SUM(cm.drv_score)                                              AS for_pts,
        SUM(opp.drv_score)                                             AS against_pts,
        SUM(cm.drv_premiership_points)                                 AS pp
    FROM afl.club_match cm
    JOIN afl.match m       ON cm.match_id = m.id
    JOIN afl.round r       ON m.round_id = r.id
    JOIN afl.season s      ON r.season_id = s.id
    JOIN afl.club_match opp ON opp.match_id = m.id AND opp.id != cm.id
    JOIN afl.club_season cs2 ON cm.club_season_id = cs2.id
    WHERE m.data_status = 'final'
    AND s.name = 'AFL 2026'
    GROUP BY cm.club_season_id
) stats
WHERE cs.id = stats.club_season_id;

-- ============================================================
-- FFL: Fix player_match status (seed used 'played'; domain uses 'named')
-- ============================================================

UPDATE ffl.player_match
SET status = 'named'
WHERE status = 'played';

-- ============================================================
-- FFL: Link player_match to AFL stats and set drv_afl_status
-- ============================================================

-- Players who have AFL Round 1 stats → 'played'
UPDATE ffl.player_match
SET
    drv_afl_status      = 'played',
    afl_player_match_id = linked.apm_id
FROM (
    SELECT fpm.id AS fpm_id, apm.id AS apm_id
    FROM ffl.player_match fpm
    JOIN ffl.player_season fps ON fpm.player_season_id = fps.id
    JOIN afl.player_season aps ON aps.id = fps.afl_player_season_id
    JOIN afl.player_match  apm ON apm.player_season_id = aps.id
    JOIN afl.club_match    acm ON apm.club_match_id = acm.id
    JOIN afl.match         am  ON acm.match_id = am.id
    JOIN afl.round         ar  ON am.round_id = ar.id
    JOIN ffl.club_match    fcm ON fpm.club_match_id = fcm.id
    JOIN ffl.match         fm  ON fcm.match_id = fm.id
    JOIN ffl.round         fr  ON fm.round_id = fr.id
    JOIN ffl.season        fs  ON fr.season_id = fs.id
    WHERE ar.name = 'Round 1'
    AND am.data_status = 'final'
    AND fs.name = 'FFL 2026'
    AND fr.name = 'Round 1'
) linked
WHERE ffl.player_match.id = linked.fpm_id;

-- Players without AFL Round 1 stats → 'dnp'
UPDATE ffl.player_match fpm
SET drv_afl_status = 'dnp'
WHERE fpm.drv_afl_status IS NULL
AND fpm.club_match_id IN (
    SELECT fcm.id FROM ffl.club_match fcm
    JOIN ffl.match  fm ON fcm.match_id = fm.id
    JOIN ffl.round  fr ON fm.round_id = fr.id
    JOIN ffl.season fs ON fr.season_id = fs.id
    WHERE fs.name = 'FFL 2026' AND fr.name = 'Round 1'
);

-- ============================================================
-- FFL: Set club_match data_status = 'final' for Round 1
-- ============================================================

UPDATE ffl.club_match fcm
SET data_status = 'final'
WHERE fcm.id IN (
    SELECT fcm2.id FROM ffl.club_match fcm2
    JOIN ffl.match  fm ON fcm2.match_id = fm.id
    JOIN ffl.round  fr ON fm.round_id = fr.id
    JOIN ffl.season fs ON fr.season_id = fs.id
    WHERE fs.name = 'FFL 2026' AND fr.name = 'Round 1'
);

-- ============================================================
-- FFL: Compute club_match premiership points and match result
-- ============================================================

UPDATE ffl.club_match fcm
SET drv_premiership_points = CASE
    WHEN fcm.drv_score > opp.drv_score THEN 4
    WHEN fcm.drv_score < opp.drv_score THEN 0
    ELSE 2
END
FROM ffl.club_match opp
WHERE opp.match_id = fcm.match_id
AND opp.id != fcm.id
AND fcm.data_status = 'final';

UPDATE ffl.match fm
SET drv_result = CASE
    WHEN home_cm.drv_score > away_cm.drv_score THEN 'home_win'
    WHEN away_cm.drv_score > home_cm.drv_score THEN 'away_win'
    ELSE 'draw'
END
FROM ffl.club_match home_cm,
     ffl.club_match away_cm
WHERE home_cm.match_id = fm.id AND home_cm.side = 'home'
AND   away_cm.match_id = fm.id AND away_cm.side = 'away'
AND home_cm.data_status = 'final';

-- ============================================================
-- FFL: Update club_season ladder stats
-- ============================================================

UPDATE ffl.club_season fcs
SET
    drv_played             = stats.played,
    drv_won                = stats.won,
    drv_lost               = stats.lost,
    drv_drawn              = stats.drawn,
    drv_for                = stats.for_pts,
    drv_against            = stats.against_pts,
    drv_premiership_points = stats.pp
FROM (
    SELECT
        fcm.club_season_id,
        COUNT(*)                                                        AS played,
        COUNT(*) FILTER (WHERE fcm.drv_premiership_points = 4)         AS won,
        COUNT(*) FILTER (WHERE fcm.drv_premiership_points = 0)         AS lost,
        COUNT(*) FILTER (WHERE fcm.drv_premiership_points = 2)         AS drawn,
        SUM(fcm.drv_score)                                              AS for_pts,
        SUM(opp.drv_score)                                             AS against_pts,
        SUM(fcm.drv_premiership_points)                                 AS pp
    FROM ffl.club_match fcm
    JOIN ffl.club_season fcs2 ON fcm.club_season_id = fcs2.id
    JOIN ffl.season      fs   ON fcs2.season_id = fs.id
    JOIN ffl.club_match  opp  ON opp.match_id = fcm.match_id AND opp.id != fcm.id
    WHERE fcm.data_status = 'final'
    AND fs.name = 'FFL 2026'
    GROUP BY fcm.club_season_id
) stats
WHERE fcs.id = stats.club_season_id;

COMMIT;
