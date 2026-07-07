# Roadmap

Committed phases — active and next. Completed phases 1–21 are in `plans/history.md`.

---

## Phase 23: UX — Player Intelligence

**Goal:** Richer player and club context for FFL decision-making. Candidate areas from `plans/ideas.md`: player season pages, free agents, club pages, Team Builder analytics, notes in the match view, score reconciliation. Scope confirmed at sprint start.

## Phase 24: Data Management — Data Setup & Historical Import

**Goal:** Season setup tooling and one-time historical backfill — the operations needed once per season (or once ever) rather than every round.

- [ ] AFL season player import — once/season CLI; fuzzy name matching to existing players; accept/reject flow for new and retiring players
- [ ] FFL squad import — once/season CLI; resolve FFL rosters to AFL player IDs
- [ ] AFL historical data import — one-time CLI from afltables CSV (2024-present already seeded; earlier years TBD)
- [ ] Pluggable FFL scoring formula — strategy pattern keyed by season; `ScoringStrategy` interface + concrete implementations covering known formula variants; `ffl.season.scoring_strategy` column; wire into score calculation use case (deferred from Phase 20)
- [ ] FFL historical team backfill — one-time CLI using `ForumPostParser` + `ImportRoundTeams` over historical forum data (requires pluggable scoring formula above)

## Phase 25: Event Reliability (REV-1)

**Goal:** Event processing fails loudly and recoverably — remove the silent-permanent-death mode found by the 2026-07 review (`doc/review-findings.md` §6, the one `critical` finding).

- [ ] `shared/events/pg`: `Listen` reconnects with backoff on error (or the service exits fatally) — never log-and-die silently
- [ ] Re-`LISTEN` on the new connection after reconnect; log recovery at warn level
- [ ] Publish failures escalate beyond a debug/warn line (error log at minimum, visible in service output)
- [ ] Verify FFL and Search listener goroutines (`ffl/cmd/main.go`, `search/cmd/main.go`) surface listener death — no warn-and-continue
- [ ] Integration test: kill the listener connection mid-run; assert events resume after reconnect
