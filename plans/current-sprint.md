# Current Sprint — Data Loading

**Sprint goal:** Build scrapers to populate AFL stats and FFL team submissions from real sources, and backfill historical FFL data.

## Tasks

### Architecture
- [ ] ADR if any new dependency introduced (e.g. goquery, Sheets API client)
- [ ] Adapters live in `internal/infrastructure/<source>/` within the relevant service
- [ ] Entry points live in `cmd/ingest/` within the relevant service

### AFL stats integration
- [ ] Choose AFL stats source (footywire, squiggle, AFL website, etc.)
- [ ] Define `StatsProvider` outbound port in AFL application layer
- [ ] Implement secondary adapter (e.g. `FootywireStatsAdapter`) in `services/afl/internal/infrastructure/<source>/`
- [ ] Wire adapter → AFL DB writes → fires `AFL.PlayerMatchUpdated` events
- [ ] `cmd/ingest/` entry point with trigger mechanism (cron, manual CLI, or live-round polling)

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