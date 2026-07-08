# afl-historical

Historical AFL data for the Phase 24 backfill (see `services/afl/cmd/`).

- `<season>.csv` — per-season player match stats scraped from afltables.com by
  `afltables-export`, and read by `afltables-import` into the AFL DB.
- `export.log` — output of the export scrape (per-season row/game counts).
- `import-review.log` — audit of the import's auto-decisions (auto-linked season
  gaps and fuzzy name near-misses).
- `gap-review.md` — outcomes of the same-name season-gap review (which players
  were split vs kept, with narratives).
