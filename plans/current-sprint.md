# Current Sprint — Data Loading

**Sprint goal:** Build scrapers to populate AFL stats and FFL team submissions from real sources, and backfill historical FFL data.

## Tasks

### Architecture
- [x] ADR-016: ACL identity mapping tables in each service's own schema (e.g. `afl.player_source_map`); no shared integration schema
- [ ] ADR if any new dependency introduced (e.g. Sheets API client)
- [ ] Adapters live in `internal/infrastructure/<source>/` within the relevant service
- [ ] Entry points live in `cmd/ingest/` within the relevant service

### AFL stats integration
- Source: `https://afltables.com/afl/stats/{year}_stats.txt` — plain CSV, one row per player per match
- No new dependency needed — Go stdlib `encoding/csv` is sufficient (no ADR required)
- Cache policy: fetch at most once per week; cache clears Monday (respectful of host)
- [ ] Add `integrations.player_source_map` table (source, external_id, player_id) — per ADR-016
- [ ] Build team code → club ID mapping (2-letter codes e.g. `SY`, `CA`)
- [ ] Define `StatsProvider` outbound port in AFL application layer
- [ ] Implement `AFLTablesAdapter` in `services/afl/internal/infrastructure/afltables/`
  - Fetch + cache CSV (weekly, cache-bust on Monday)
  - Parse rows → domain PlayerMatch stats
  - Upsert by `(afl_player_id, round)` for idempotency
- [ ] Wire adapter → AFL DB writes → fires `AFL.PlayerMatchUpdated` events
- [ ] `cmd/ingest/` entry point (manual trigger; can be put on a cron later)

### FFL team submissions integration
- [ ] Identify nominated sources (Google Sheets, email, form, etc.)
- [ ] Define `TeamSubmissionSource` outbound port in FFL application layer
- [ ] Implement secondary adapter for each source in `services/ffl/internal/infrastructure/<source>/`
- [ ] Wire adapter → FFL DB writes (PlayerMatch, ClubMatch rows per round)
- [ ] Handle re-submission / idempotency
- [ ] `cmd/ingest/` entry point

### Historical FFL data (one-time)
- [ ] Identify where historical data lives and its format
- [ ] Implement adapter + use case (entry point under `dev/` — migration tool, not production binary)
- [ ] Run against dev, then prod
- [ ] Verify ladder, scores, and player history look correct post-import

### Database backup
- [ ] Choose cloud backup location (e.g. S3, GCS, Backblaze B2)
- [ ] Script `pg_dump` → compress → upload to chosen location
- [ ] Verify restore works from backup
- [ ] Run backup after all data is loaded and verified