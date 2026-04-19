# Current Sprint — Phase 14: Historical AFL Data

**Sprint goal:** Load real AFL player match stats (2024–present) from afltables.com into the domain. Establishes the canonical player roster and the reconciliation tooling all future imports will reuse.

See `ai/architecture/historical-import.md` for the two-phase pattern, xref conventions, and fuzzy matching strategy.

## Tasks

### Architecture
- [x] ADR-016: xref tables per source per entity; no FK to core schema; no shared integration schema

### AFL historical stats import
- Source: `https://afltables.com/afl/stats/{year}_stats.txt` — plain CSV, one row per player per match
- Entry point: `dev/import/afl_historical/main.go` — dev tool, not a production binary
- Two-phase: `--reconcile` outputs `reconcile.csv` for human review; default run imports using that file
- xref table `afl.xref_afltables_player` (external_id → afl.player.id) is the durable identity mapping
- No events fired; no production caching needed
- [x] Build team code → club name mapping (`services/afl/internal/infrastructure/afltables/clubs.go`)
- [x] Build CSV parser (`services/afl/internal/infrastructure/afltables/parser.go`)
- [ ] Restore `afl.xref_afltables_player` table — `dev/postgres/init/03_afl_integrations.sql`
- [ ] Build `dev/import/afl_historical/main.go` — `--reconcile` and import modes; Levenshtein player matching
- [ ] Run `--reconcile` against dev DB; review and complete `reconcile.csv`; commit it
- [ ] Import into dev DB; verify player/match counts and spot-check stats
- [ ] Run against prod DB (wipe 2026 fake fixture first)

## Up next
- Phase 15: Database backup — `pg_dump` → compress → upload; verify restore
- Phase 16: Historical FFL data import