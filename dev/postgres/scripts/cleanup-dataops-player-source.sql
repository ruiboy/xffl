-- RAN on 2026-06-08 — safe to delete this file from 2026-06-15 onwards.
--
-- Cleanup: dataops_player_source rows written for EXACT name matches.
--
-- Before the fix in ResolveAFLPlayerMatch, every "Link Player" action wrote a
-- afl.dataops_player_source mapping row, even when the footywire-parsed name
-- matched the linked player's actual name exactly. Those rows aren't needed —
-- addAFLPlayerSeason already ensures a current-season player_season exists, so
-- the exact-name match will resolve via the normal candidate pool next import.
--
-- This identifies the 71 rows (of 112 total, captured 2026-06-08) from the
-- 2026-06-07 AFL stats import where external_player matched the linked
-- player's name case-insensitively, and removes them.

-- Preview before deleting:
SELECT dps.source, dps.external_season, dps.external_club, dps.external_player,
       dps.player_season_id, p.name AS actual_player_name, dps.created_at
FROM afl.dataops_player_source dps
JOIN afl.player_season ps ON ps.id = dps.player_season_id
JOIN afl.player p ON p.id = ps.player_id
WHERE lower(trim(dps.external_player)) = lower(trim(p.name))
ORDER BY dps.external_club, dps.external_player;

-- Delete the exact-match rows:
BEGIN;
DELETE FROM afl.dataops_player_source dps
USING afl.player_season ps, afl.player p
WHERE dps.player_season_id = ps.id
  AND ps.player_id = p.id
  AND lower(trim(dps.external_player)) = lower(trim(p.name));
ROLLBACK;
COMMIT;

-- ============================================================================
-- RESTORE — run this if it turns out these rows were needed after all.
-- Snapshot taken 2026-06-08, before deletion. Re-running is safe (upsert on PK).
-- ============================================================================

INSERT INTO afl.dataops_player_source
    (source, external_season, external_club, external_player, player_season_id, created_at, updated_at)
VALUES
    ('footywire', '1', '1', 'Charlie Edwards', 1895, '2026-06-07 14:37:57.308376+00', '2026-06-07 14:37:57.308376+00'),
    ('footywire', '1', '1', 'Daniel Curtin', 1919, '2026-06-07 15:04:10.258713+00', '2026-06-07 15:04:10.258713+00'),
    ('footywire', '1', '10', 'Bailey Macdonald', 1946, '2026-06-07 15:17:46.637573+00', '2026-06-07 15:17:46.637573+00'),
    ('footywire', '1', '10', 'Bodie Ryan', 1933, '2026-06-07 15:08:13.811119+00', '2026-06-07 15:08:13.811119+00'),
    ('footywire', '1', '10', 'Calsher Dear', 1892, '2026-06-07 14:35:32.255328+00', '2026-06-07 14:35:32.255328+00'),
    ('footywire', '1', '10', 'Cameron Nairn', 1947, '2026-06-07 15:18:09.003129+00', '2026-06-07 15:18:09.003129+00'),
    ('footywire', '1', '10', 'Henry Hustwaite', 1940, '2026-06-07 15:14:28.363394+00', '2026-06-07 15:14:28.363394+00'),
    ('footywire', '1', '10', 'Max Ramsden', 1934, '2026-06-07 15:08:21.781373+00', '2026-06-07 15:08:21.781373+00'),
    ('footywire', '1', '10', 'Will Day', 1954, '2026-06-07 15:22:02.654957+00', '2026-06-07 15:22:02.654957+00'),
    ('footywire', '1', '10', 'William McCabe', 1893, '2026-06-07 14:36:30.535934+00', '2026-06-07 14:36:30.535934+00'),
    ('footywire', '1', '11', 'Jake Bowey', 1925, '2026-06-07 15:05:51.863439+00', '2026-06-07 15:05:51.863439+00'),
    ('footywire', '1', '11', 'Matthew Jefferson', 1899, '2026-06-07 14:40:52.157883+00', '2026-06-07 14:40:52.157883+00'),
    ('footywire', '1', '11', 'Max Heath', 1901, '2026-06-07 14:42:09.603129+00', '2026-06-07 14:42:09.603129+00'),
    ('footywire', '1', '11', 'Xavier Taylor', 1900, '2026-06-07 14:41:30.232292+00', '2026-06-07 14:41:30.232292+00'),
    ('footywire', '1', '12', 'Riley Hardeman', 1948, '2026-06-07 15:19:05.835272+00', '2026-06-07 15:19:05.835272+00'),
    ('footywire', '1', '12', 'Wil Dawson', 1931, '2026-06-07 15:07:22.014385+00', '2026-06-07 15:07:22.014385+00'),
    ('footywire', '1', '12', 'Zac Fisher', 1960, '2026-06-07 15:34:01.692316+00', '2026-06-07 15:34:01.692316+00'),
    ('footywire', '1', '13', 'Christian Moraes', 1894, '2026-06-07 14:36:50.999907+00', '2026-06-07 14:36:50.999907+00'),
    ('footywire', '1', '13', 'Tom Anastasopoulos', 1942, '2026-06-07 15:15:06.136621+00', '2026-06-07 15:15:06.136621+00'),
    ('footywire', '1', '13', 'Tom Cochrane', 1968, '2026-06-07 16:39:17.695893+00', '2026-06-07 16:39:17.695893+00'),
    ('footywire', '1', '13', 'Will Lorenz', 1941, '2026-06-07 15:14:46.44363+00', '2026-06-07 15:14:46.44363+00'),
    ('footywire', '1', '14', 'Patrick Retschko', 1896, '2026-06-07 14:39:33.08442+00', '2026-06-07 14:39:33.08442+00'),
    ('footywire', '1', '14', 'Sam Cumming', 1906, '2026-06-07 14:46:50.148234+00', '2026-06-07 14:46:50.148234+00'),
    ('footywire', '1', '14', 'Thomas Burton', 1907, '2026-06-07 14:57:36.538241+00', '2026-06-07 14:57:36.538241+00'),
    ('footywire', '1', '15', 'Angus Hastie', 1964, '2026-06-07 15:36:52.172276+00', '2026-06-07 15:36:52.172276+00'),
    ('footywire', '1', '15', 'Campbell Lake', 1963, '2026-06-07 15:36:38.391585+00', '2026-06-07 15:36:38.391585+00'),
    ('footywire', '1', '15', 'Charlie Banfield', 1937, '2026-06-07 15:12:16.007399+00', '2026-06-07 15:12:16.007399+00'),
    ('footywire', '1', '15', 'Liam Henry', 1915, '2026-06-07 15:02:43.570184+00', '2026-06-07 15:02:43.570184+00'),
    ('footywire', '1', '15', 'Ryan Byrnes', 1927, '2026-06-07 15:06:18.259281+00', '2026-06-07 15:06:18.259281+00'),
    ('footywire', '1', '16', 'Billy Cootee', 1913, '2026-06-07 14:59:54.079245+00', '2026-06-07 14:59:54.079245+00'),
    ('footywire', '1', '16', 'Harry Kyle', 1957, '2026-06-07 15:23:07.14631+00', '2026-06-07 15:23:07.14631+00'),
    ('footywire', '1', '16', 'Hayden McLean', 1930, '2026-06-07 15:06:59.475627+00', '2026-06-07 15:06:59.475627+00'),
    ('footywire', '1', '16', 'Peter Ladhams', 1956, '2026-06-07 15:22:49.700064+00', '2026-06-07 15:22:49.700064+00'),
    ('footywire', '1', '16', 'Tom Hanily', 1949, '2026-06-07 15:19:45.37613+00', '2026-06-07 15:19:45.37613+00'),
    ('footywire', '1', '16', 'Will Edwards', 1912, '2026-06-07 14:59:35.894825+00', '2026-06-07 14:59:35.894825+00'),
    ('footywire', '1', '17', 'Brandon Starcevich', 1967, '2026-06-07 16:39:09.052219+00', '2026-06-07 16:39:09.052219+00'),
    ('footywire', '1', '17', 'Harvey Johnston', 1922, '2026-06-07 15:05:15.800945+00', '2026-06-07 15:05:15.800945+00'),
    ('footywire', '1', '17', 'Jack Hutchinson', 1923, '2026-06-07 15:05:25.561299+00', '2026-06-07 15:05:25.561299+00'),
    ('footywire', '1', '17', 'Jack Williams', 1903, '2026-06-07 14:45:11.025192+00', '2026-06-07 14:45:11.025192+00'),
    ('footywire', '1', '17', 'Rhett Bazzo', 1936, '2026-06-07 15:11:25.065365+00', '2026-06-07 15:11:25.065365+00'),
    ('footywire', '1', '18', 'Adam Treloar', 1908, '2026-06-07 14:58:23.983059+00', '2026-06-07 14:58:23.983059+00'),
    ('footywire', '1', '18', 'Cody Weightman', 1969, '2026-06-07 16:42:53.548748+00', '2026-06-07 16:42:53.548748+00'),
    ('footywire', '1', '18', 'Jedd Busslinger', 1909, '2026-06-07 14:58:31.558453+00', '2026-06-07 14:58:31.558453+00'),
    ('footywire', '1', '18', 'Lachlan Smith', 1910, '2026-06-07 14:58:55.823107+00', '2026-06-07 14:58:55.823107+00'),
    ('footywire', '1', '18', 'Laitham Vandermeer', 1921, '2026-06-07 15:04:43.234068+00', '2026-06-07 15:04:43.234068+00'),
    ('footywire', '1', '18', 'Luke Cleary', 1953, '2026-06-07 15:21:25.652624+00', '2026-06-07 15:21:25.652624+00'),
    ('footywire', '1', '18', 'Ryan Gardner', 1911, '2026-06-07 14:59:05.860485+00', '2026-06-07 14:59:05.860485+00'),
    ('footywire', '1', '2', 'Cody Curtin', 1918, '2026-06-07 15:03:52.058277+00', '2026-06-07 15:03:52.058277+00'),
    ('footywire', '1', '2', 'Conor McKenna', 1902, '2026-06-07 14:43:06.325533+00', '2026-06-07 14:43:06.325533+00'),
    ('footywire', '1', '2', 'Sam Marshall', 1945, '2026-06-07 15:15:37.388546+00', '2026-06-07 15:15:37.388546+00'),
    ('footywire', '1', '2', 'Shadeau Brain', 1962, '2026-06-07 15:35:26.240425+00', '2026-06-07 15:35:26.240425+00'),
    ('footywire', '1', '2', 'Will McLachlan', 1952, '2026-06-07 15:21:00.36231+00', '2026-06-07 15:21:00.36231+00'),
    ('footywire', '1', '3', 'Billy Wilson', 1932, '2026-06-07 15:07:40.68096+00', '2026-06-07 15:07:40.68096+00'),
    ('footywire', '1', '3', 'Blake Acres', 1955, '2026-06-07 15:22:27.602037+00', '2026-06-07 15:22:27.602037+00'),
    ('footywire', '1', '3', 'Flynn Young', 1950, '2026-06-07 15:20:23.292352+00', '2026-06-07 15:20:23.292352+00'),
    ('footywire', '1', '3', 'Jack Ison', 1939, '2026-06-07 15:14:16.974119+00', '2026-06-07 15:14:16.974119+00'),
    ('footywire', '1', '4', 'Charlie West', 1944, '2026-06-07 15:15:26.214581+00', '2026-06-07 15:15:26.214581+00'),
    ('footywire', '1', '4', 'Harvey Harrison', 1958, '2026-06-07 15:23:44.44535+00', '2026-06-07 15:23:44.44535+00'),
    ('footywire', '1', '5', 'Nick Bryan', 1929, '2026-06-07 15:06:39.648628+00', '2026-06-07 15:06:39.648628+00'),
    ('footywire', '1', '6', 'Jeremy Sharp', 1904, '2026-06-07 14:45:37.067088+00', '2026-06-07 14:45:37.067088+00'),
    ('footywire', '1', '6', 'Mason Cox', 1897, '2026-06-07 14:40:03.374737+00', '2026-06-07 14:40:03.374737+00'),
    ('footywire', '1', '6', 'Michael Frederick', 1898, '2026-06-07 14:40:22.689871+00', '2026-06-07 14:40:22.689871+00'),
    ('footywire', '1', '7', 'Rhys Stanley', 1926, '2026-06-07 15:06:04.354198+00', '2026-06-07 15:06:04.354198+00'),
    ('footywire', '1', '8', 'Ben Jepson', 1928, '2026-06-07 15:06:28.890781+00', '2026-06-07 15:06:28.890781+00'),
    ('footywire', '1', '8', 'Jai Murray', 1961, '2026-06-07 15:35:13.161261+00', '2026-06-07 15:35:13.161261+00'),
    ('footywire', '1', '8', 'Ned Moyle', 1914, '2026-06-07 15:02:16.950185+00', '2026-06-07 15:02:16.950185+00'),
    ('footywire', '1', '9', 'Conor Stone', 1951, '2026-06-07 15:20:44.695749+00', '2026-06-07 15:20:44.695749+00'),
    ('footywire', '1', '9', 'Harrison Oliver', 1935, '2026-06-07 15:10:40.565036+00', '2026-06-07 15:10:40.565036+00'),
    ('footywire', '1', '9', 'Leek Aleer', 1890, '2026-06-07 14:33:29.34235+00', '2026-06-07 14:33:29.34235+00'),
    ('footywire', '1', '9', 'Sam Taylor', 1959, '2026-06-07 15:24:08.667434+00', '2026-06-07 15:24:08.667434+00'),
    ('footywire', '1', '9', 'Toby McMullin', 1891, '2026-06-07 14:33:47.959749+00', '2026-06-07 14:33:47.959749+00')
ON CONFLICT (source, external_season, external_club, external_player) DO NOTHING;
