# Roadmap

## Context

Full stack rebuild (backend + frontend). Gateway introduced early so frontends always connect through it. All frontend phases require Playwright tests. See `plans/history.md` for completed phases 1–21.

## Phase 22: Real Data Load & Seed Cleanup

**Goal:** Replace synthetic seed data with real 2026 AFL and FFL data, and slim the dev seed to a lean scaffold.

- [ ] Import all 2026 AFL rounds played to date (teams, players, match stats) via existing import tooling
- [ ] Import all 2026 FFL round teams to date
- [ ] Verify ladder calculation produces correct standings
- [ ] Smoke-test team submission and substitution flows against real data; capture edge cases
- [ ] Reduce `dev/seed` to a single representative round (enough for demo and future dev)
- [ ] Confirm e2e tests pass against slim seed

## Phase 23: UX — Player Intelligence

**Goal:** Richer player and club context for FFL decision-making. Candidate areas from `plans/ideas.md`: player season pages, free agents, club pages, Team Builder analytics, notes in the match view, score reconciliation. Scope confirmed at sprint start.

## Phase 24: AFL Player Availability

**Goal:** Track AFL injury and suspension data per player season, and enforce availability at FFL team submission.

- [ ] Add availability columns to `afl.player_season`: `availability TEXT`, `available_from_round_id INTEGER`, `availability_note TEXT`
- [ ] Update init SQL and test-e2e init SQL; apply via `ALTER TABLE` to live DB per migration strategy in cookbook
- [ ] Import pipeline: pull AFL injury/suspension list; populate availability fields
- [ ] FFL team submission validation: reject naming an unavailable player
- [ ] UX: show availability status on free agents page (PAGE-4) and player season view (PAGE-1)

## Phase 25: Data Management — Data Setup & Historical Import

**Goal:** Season setup tooling and one-time historical backfill — the operations needed once per season (or once ever) rather than every round.

- [ ] AFL season player import — once/season CLI; fuzzy name matching to existing players; accept/reject flow for new and retiring players
- [ ] FFL squad import — once/season CLI; resolve FFL rosters to AFL player IDs
- [ ] AFL historical data import — one-time CLI from afltables CSV (2024-present already seeded; earlier years TBD)
- [ ] Pluggable FFL scoring formula — strategy pattern keyed by season; `ScoringStrategy` interface + concrete implementations covering known formula variants; `ffl.season.scoring_strategy` column; wire into score calculation use case (deferred from Phase 20)
- [ ] FFL historical team backfill — one-time CLI using `ForumPostParser` + `ImportRoundTeams` over historical forum data (requires pluggable scoring formula above)

## Phase 26: Search Frontend + Index Enrichment

**Goal:** Search UI backed by an enriched index. Starts with a required ADR review — the technology decision must be made before building anything.

- [ ] **Decision gate:** revisit ADR-015 — Typesense (text search strength) vs ClickHouse (aggregation-heavy analytics); probe trade-offs and amend or supersede ADR-015 before committing to either implementation
- [ ] Search view — full-text search with filters (source, type)
- [ ] Expand search index to support UX data requirements (player stats, aggregates, etc.)
- [ ] Playwright tests

## Phase 27: Deployment

- [ ] CI-ready (GitHub Actions or similar)
- [ ] ADR — Consider deployment options (AWS, GCP, etc.)

---

## Future Ideas & Parked Designs

See `plans/ideas.md`.
