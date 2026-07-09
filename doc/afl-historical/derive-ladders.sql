-- Derive AFL match scores + season ladders for the historical backfill.
--
-- One-off derivation over the imported 1998–2023 data (the import writes raw
-- player_match facts only; this computes the drv_* columns). Intended to become
-- the "run all derivations" second step of the import (kept separate from the
-- scrape+write machinery). Re-runnable and idempotent.
--
-- Validated: run against 2026 (finals-excluded) it reproduces the domain-computed
-- (CalculateLadder) club_season standings exactly — so this SQL matches the Go
-- ladder logic. NB: the ladder MUST exclude finals rounds (they don't count
-- toward the home-and-away ladder); the Go CalculateLadder/FindFinalBySeasonID
-- path does not yet do this (only safe today because 2026 has no finals) — the
-- future derive step needs the same `NOT ILIKE '%Final%'` filter.
--
-- Scope defaults to 1998–2023; adjust the year range as needed.

\set from 1998
\set to 2023

-- 1) club_match.drv_score = goals*6 + behinds + rushed_behinds (per club, per match)
UPDATE afl.club_match cm
SET drv_score = COALESCE((SELECT SUM(pm.goals)*6 + SUM(pm.behinds)
                          FROM afl.player_match pm WHERE pm.club_match_id = cm.id), 0)
                + cm.rushed_behinds,
    updated_at = CURRENT_TIMESTAMP
FROM afl.match m
JOIN afl.round r ON r.id = m.round_id
JOIN afl.season s ON s.id = r.season_id
WHERE cm.match_id = m.id
  AND m.data_status = 'final'
  AND regexp_replace(s.name, '\D', '', 'g')::int BETWEEN :from AND :to;

-- 2) club_season ladder standings from home-and-away results (finals excluded).
--    Mirrors domain.CalculateLadder: 4 pts a win, 2 a draw, plus for/against.
WITH mr AS (
  SELECT home.club_season_id AS cs, home.drv_score AS f, away.drv_score AS a
  FROM afl.match m
  JOIN afl.club_match home ON home.match_id = m.id AND home.side = 'home'
  JOIN afl.club_match away ON away.match_id = m.id AND away.side = 'away'
  JOIN afl.round r ON r.id = m.round_id
  JOIN afl.season s ON s.id = r.season_id
  WHERE m.data_status = 'final'
    AND regexp_replace(s.name, '\D', '', 'g')::int BETWEEN :from AND :to
    AND r.name NOT ILIKE '%Final%'
  UNION ALL
  SELECT away.club_season_id, away.drv_score, home.drv_score
  FROM afl.match m
  JOIN afl.club_match home ON home.match_id = m.id AND home.side = 'home'
  JOIN afl.club_match away ON away.match_id = m.id AND away.side = 'away'
  JOIN afl.round r ON r.id = m.round_id
  JOIN afl.season s ON s.id = r.season_id
  WHERE m.data_status = 'final'
    AND regexp_replace(s.name, '\D', '', 'g')::int BETWEEN :from AND :to
    AND r.name NOT ILIKE '%Final%'
),
agg AS (
  SELECT cs,
    COUNT(*)                       AS played,
    COUNT(*) FILTER (WHERE f > a)  AS won,
    COUNT(*) FILTER (WHERE f < a)  AS lost,
    COUNT(*) FILTER (WHERE f = a)  AS drawn,
    SUM(f)                         AS for_pts,
    SUM(a)                         AS against_pts,
    4 * COUNT(*) FILTER (WHERE f > a) + 2 * COUNT(*) FILTER (WHERE f = a) AS pp
  FROM mr GROUP BY cs
)
UPDATE afl.club_season c
SET drv_played = agg.played, drv_won = agg.won, drv_lost = agg.lost, drv_drawn = agg.drawn,
    drv_for = agg.for_pts, drv_against = agg.against_pts, drv_premiership_points = agg.pp,
    updated_at = CURRENT_TIMESTAMP
FROM agg WHERE c.id = agg.cs;
