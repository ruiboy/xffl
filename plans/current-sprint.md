# Current Sprint — Phase 15: Database Backup

**Sprint goal:** Persist DB state durably outside the dev lifecycle. Now that real AFL historical data is seeded, it must survive `dev-reset`, machine loss, or accidental wipes. Backup runs on demand and on a regular schedule.

## Tasks

### Infrastructure
- [x] `just backup-db` recipe — `pg_dump | gzip` → timestamped `backups/postgres_YYYYMMDD_HHMMSS.sql.gz`; uploads via rclone if `BACKUP_REMOTE` is set
- [x] Verify backup end-to-end: `just backup` → timestamped `.sql.gz` (292K), no warnings
- [x] Verify restore: `dev-reset` → `dev-up` → restore → row counts match (810 players, 21942 player_match, 30 ffl players)
- [x] Upload deferred — backup writes locally for now; cloud destination is a future item

## Up next
- Phase 16: Historical FFL data import