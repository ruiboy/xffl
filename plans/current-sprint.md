# Current Sprint — Phase 16: 2026 FFL Data Import

**Sprint goal:** Seed real 2026 FFL data into the DB — squad rosters, round team selections, and match scores for R1–R5.

## Tasks

- [x] Identify data source — Tapatalk forum posts (manual copy-paste)
- [x] Build `dev/import/ffl/parse_forum.py` — parses all 4 team formats → `*_teams.csv` + `*_scores.csv`
- [x] Parse and validate R1–R6 (88 player rows/round); R6 squads parsed (no scores)
- [x] Build `dev/import/ffl/resolve_squads.py` — matches squad players to `afl.player` records, generates SQL
- [x] Generate and apply `dev/postgres/seed/04_ffl_players.sql` — 120 players (30 × 4 clubs) with player_seasons
- [x] Build `dev/import/ffl/import_round_teams.py` — matches CSV players to player_seasons, generates/applies player_match + club_match score SQL
- [x] Add 4 mid-season trades to `dev/postgres/seed/05_ffl_trades.sql` (Houston, Davies-Uniacke, Visentini, Josh Treacy)
- [x] Insert round team selections (ffl.player_match) + scores (ffl.club_match) for R1–R5 — 437 statements applied, 22 players/team across all rounds
- [ ] Verify ladder standings and scores post-import

## Up next
- Phase 17: UX Improvements