# Current Sprint — Data Loading

**Sprint goal:** Build scrapers to populate AFL stats and FFL team submissions from real sources, and backfill historical FFL data.

## Tasks

### Architecture
- [x] ADR-016: ACL identity mapping tables in each service's own schema; no shared integration schema
- [ ] ADR if any new dependency introduced (e.g. Sheets API client)

### AFL historical stats import (one-time dev tool)
- Source: `https://afltables.com/afl/stats/{year}_stats.txt` — plain CSV, one row per player per match; covers 2024–present
- Entry point: `dev/import/afl_historical/main.go` — not a production binary (see `ai/architecture/historical-import.md`)
- Two-phase: `--reconcile` outputs `dev/import/afl_historical/reconcile.csv` for human review; default run imports using that file
- xref table `afl.xref_afltables_player` (external_id → afl.player.id) is the durable identity mapping
- No events fired; no production caching needed
- [x] Build team code → club name mapping (`services/afl/internal/infrastructure/afltables/clubs.go`)
- [x] Build CSV parser (`services/afl/internal/infrastructure/afltables/parser.go`)
- [ ] Restore `afl.xref_afltables_player` table — `dev/postgres/init/03_afl_integrations.sql`
- [ ] Build `dev/import/afl_historical/main.go` — `--reconcile` mode and import mode
- [ ] Run `--reconcile` against dev DB; review and complete `reconcile.csv`; commit it
- [ ] Import into dev DB; verify player/match counts and spot-check stats
- [ ] Run against prod DB (wipe 2026 fake data first)

### FFL team submissions integration
- [ ] Identify nominated sources (Google Sheets, email, form, etc.)
- [ ] Define `TeamSubmissionSource` outbound port in FFL application layer
- [ ] Implement secondary adapter for each source in `services/ffl/internal/infrastructure/<source>/`
- [ ] Wire adapter → FFL DB writes (PlayerMatch, ClubMatch rows per round)
- [ ] Handle re-submission / idempotency
- [ ] `cmd/ingest/` entry point

### Historical FFL data (one-time)
- [ ] Identify where historical data lives and its format
- [ ] Implement adapter + use case (entry point at `dev/import/ffl_historical/main.go` — migration tool, not production binary)
- [ ] Run against dev, then prod
- [ ] Verify ladder, scores, and player history look correct post-import

### Database backup
- [ ] Choose cloud backup location (e.g. S3, GCS, Backblaze B2)
- [ ] Script `pg_dump` → compress → upload to chosen location
- [ ] Verify restore works from backup
- [ ] Run backup after all data is loaded and verified