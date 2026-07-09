# afl-historical (archive)

Provenance archive of the Phase 24 AFL historical backfill (see
`services/afl/cmd/`). The import/export CLIs still write fresh output to a
working `afl-historical/` at the repo root; this is a one-off snapshot kept for
the record.

- `<season>.csv` — per-season player match stats scraped from afltables.com by
  `afltables-export`, and read by `afltables-import` into the AFL DB.
- `export.log` — output of the export scrape (per-season row/game counts).
- `import-review.log` — audit of the import's auto-decisions (auto-linked season
  gaps and fuzzy name near-misses).
- `gap-review.md` — outcomes of the same-name season-gap review (which players
  were split vs kept, with narratives).
- `rushed-behinds.log` — rushed-behinds backfill (team behinds − player behinds).
- `derive-ladders.sql` — reusable SQL to derive AFL ladder