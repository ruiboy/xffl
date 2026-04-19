# Current Sprint — Phase 14: Historical AFL Data

**Sprint goal:** Seed real AFL player match stats (2024–2026) from afltables.com CSV files into the domain. Establishes the canonical player roster that FFL imports will reconcile against.

## Tasks

### Architecture
- [x] ADR-016: xref tables per source per entity; no FK to core schema; no shared integration schema *(superseded — see below)*
- [x] Remove premature `ai/architecture/historical-import.md` and dead afltables package

### AFL historical stats import
- Source: 3 CSV files from afltables.com (2024, 2025, 2026)
- Approach: generate SQL seed file directly from CSVs — no import tool needed
- Same player name across years = one `afl.player` row; club changes handled by `afl.player_season`
- 2026: player stats for completed rounds only — existing 2026 fixture structure is valid, keep it
- [ ] Generate `dev/postgres/seed/03_afl_historical.sql` from the 3 CSV files
- [ ] Load into dev DB; verify player/match counts and spot-check stats
- [ ] Run against prod DB

## Up next
- Phase 15: Database backup — `pg_dump` → compress → upload; verify restore
- Phase 16: Historical FFL data import