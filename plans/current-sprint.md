# Current Sprint — Phase 24: Data Management — Data Setup & Historical Import

**Sprint goal:** Season setup tooling and one-time historical backfill — the operations needed once per season (or once ever) rather than every round.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [x] AFL historical data import — `afltables-export` (scrape → `afl-historical/<season>.csv`) + `afltables-import` (CSV → DB) CLIs; loaded 1998–2023 (249,870 player-matches) into dev; season-gap heuristic + club dedup + fuzzy handling; 9 same-name wrong-merges reviewed and split (see `afl-historical/gap-review.md`)
- [ ] Pluggable FFL scoring formula — strategy pattern keyed by season; `ScoringStrategy` interface + concrete implementations covering known formula variants; `ffl.season.scoring_strategy` column; wire into score calculation use case (deferred from Phase 20)
- [ ] FFL historical team backfill — one-time CLI using `ForumPostParser` + `ImportRoundTeams` over historical forum data (requires pluggable scoring formula above)
- [ ] AFL season player import — once/season CLI; fuzzy name matching to existing players; accept/reject flow for new and retiring players
- [ ] FFL squad import — once/season CLI; resolve FFL rosters to AFL player IDs
