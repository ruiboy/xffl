# Roadmap

Committed phases — active and next. Completed phases 1–23 are in `plans/history.md`.

---

## Phase 24: Data Management — Data Setup & Historical Import

**Goal:** Season setup tooling and one-time historical backfill — the operations needed once per season (or once ever) rather than every round.

- [x] AFL historical data import — afltables scrape + import CLIs; 1998–2023 loaded, 2024/2025 finalised + derived (scores, rushed behinds, ladders). Made navigable in the webapp (season picker, season-in-URL, ladder home) and finals typed via `round_type`.

## Phase 25: FFL Scoring & Historical Import

**Goal:** Make FFL scoring pluggable per season, then backfill historical FFL teams from forum data with era-correct scoring.

- [ ] Pluggable FFL scoring formula — strategy pattern keyed by season; `ScoringStrategy` interface + concrete implementations covering known formula variants; `ffl.season.scoring_strategy` column; wire into score calculation use case (deferred from Phase 20)
- [ ] FFL historical team backfill — one-time CLI using `ForumPostParser` + `ImportRoundTeams` over historical forum data (requires pluggable scoring formula above)

## Phase 26: Event Reliability (REV-1)

**Goal:** Event processing fails loudly and recoverably — remove the silent-permanent-death mode found by the 2026-07 review (`doc/review-findings.md` §6, the one `critical` finding).

- [ ] `shared/events/pg`: `Listen` reconnects with backoff on error (or the service exits fatally) — never log-and-die silently
- [ ] Re-`LISTEN` on the new connection after reconnect; log recovery at warn level
- [ ] Publish failures escalate beyond a debug/warn line (error log at minimum, visible in service output)
- [ ] Verify FFL and Search listener goroutines (`ffl/cmd/main.go`, `search/cmd/main.go`) surface listener death — no warn-and-continue
- [ ] Integration test: kill the listener connection mid-run; assert events resume after reconnect
