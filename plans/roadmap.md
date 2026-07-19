# Roadmap

Committed phases — active and next. Completed phases 1–24 are in `plans/history.md`.

---

## Phase 25: FFL Scoring & Historical Import

**Goal:** Make FFL scoring pluggable per season, then backfill historical FFL teams from forum data with era-correct scoring.

- [x] Pluggable FFL scoring formula — per-season `Rules` value object with `Scoring` and `Composition` facets (`ffl.season.rules_id`) + generic scoring engine; refactor-to-parity first, then historical eras defined in `rules_eras.go`. Supersedes the `ScoringStrategy` interface / `scoring_strategy` column originally scoped here — parameterised rules replaced one implementation per variant. See [ffl-scoring-rules.md](ffl-scoring-rules.md)
- [ ] FFL historical import (2006–2025) — human-in-the-loop capture (userscript → ingest → parse) feeding a layered pipeline (fixtures → squads → trades → submitted teams), with scores recomputed via per-season `Rules` and posted scores kept in notes for reconciliation. Replaces the one-time-CLI approach originally scoped. Slices 0–2 done (eras, capture, season + fixtures); 3–6 remain (squads, trades, submitted teams, coverage + reconciliation). See [ffl-historical-import.md](ffl-historical-import.md)
- [ ] Deletion semantics fast-follow (ADR-020 step 1) — `ffl.player_match.club_match_id` from `ON DELETE CASCADE` to `RESTRICT`, so a fixture delete slipping past the `hasTeams` guard fails loudly instead of silently destroying submitted teams. See [ADR-020](../ai/decisions/adr-020-deletion-semantics.md)

Emerged during the phase rather than being scoped up front — byes and superbyes were deferred from slice 2 and added as-we-go, and the rest came out of using the fixture builder:

- [x] Match model unification — one match = a style (`versus` / `bye` / `superbye`) plus its participating `club_match` rows, replacing the home/away pair. Byes and superbyes render on round and match pages; a bye counts toward For only, not games played
- [x] Fixture builder — staged round-by-round editor, round-robin generator plus manual rounds, superbyes, saves reconciled in place rather than rebuilt; rounds ordered by AFL round rather than creation order
- [x] Add Season rework — explicit rules-era selection and club picker at creation time
- [x] Admin / Data Ops separation — data-ops split into its own package and schema, Admin (Seasons, Fixtures) split out of Data Ops, Calculate tab moved to Data Ops, nav decluttered

Not-found and player navigation, arising from use rather than plan:

- [x] Not-found handling — single-entity queries resolve unknown and unparseable ids to null instead of erroring (`pgx.ErrNoRows` → `domain.ErrNotFound` through repo and resolver), a shared `NotFound` page across the FFL and AFL views, and a catch-all route for unmatched paths. Remaining unreached call paths and the typecheck gate are Phase 27
- [x] Player navigation — header player search jumping to any player's season page, and a player's other seasons linked as year + club chips from their season page

## Phase 26: Event Reliability (REV-1)

**Goal:** Event processing fails loudly and recoverably — remove the silent-permanent-death mode found by the 2026-07 review (`doc/review-findings.md` §6, the one `critical` finding).

- [ ] `shared/events/pg`: `Listen` reconnects with backoff on error (or the service exits fatally) — never log-and-die silently
- [ ] Re-`LISTEN` on the new connection after reconnect; log recovery at warn level
- [ ] Publish failures escalate beyond a debug/warn line (error log at minimum, visible in service output)
- [ ] Verify FFL and Search listener goroutines (`ffl/cmd/main.go`, `search/cmd/main.go`) surface listener death — no warn-and-continue
- [ ] Integration test: kill the listener connection mid-run; assert events resume after reconnect

## Phase 27: Lookup Error Handling & Type Gate

**Goal:** Finish the not-found work started in Phase 25 (`2fe2762`, `e677fc1`) and make the frontend typecheck a blocking gate. Nothing here is reachable from a routed page today — these are the same latent trap that produced the same bug three times, so they're closed together rather than one page at a time.

- [ ] Map `pgx.ErrNoRows` → `domain.ErrNotFound` in the nine remaining `FindByID` methods — AFL: `Club`, `ClubMatch`, `Player`, `PlayerMatch`; FFL: `Club`, `Player`, `PlayerMatch`, `PlayerSeason`, plus `ClubSeasonRepository.FindByClubAndSeason`. Mechanical, but they sit on command/loader paths, so run the suite alongside
- [ ] Replace the `err.Error() == "no rows in result set"` string comparison in AFL `PlayerSeasonRepository.FindLatestByPlayerID` with `errors.Is`
- [ ] Decide what `fflClub`, `fflPlayer` and `aflClub` do with an unresolvable id — they return raw parse errors and are non-nullable in the schema, so null needs a schema change. Only worth doing if something routes to a club or player page
- [ ] Clear the 7 `vue-tsc` errors and make the typecheck blocking in CI — 4 implicit-`any`, 1 unused variable, `SeasonBuilder.vue` calling `.at()` against a `lib: ES2020` target, and `TeamBuilderView.vue:507` comparing non-overlapping types (always false — possibly dead logic hiding a real bug). The typecheck was silently broken for a long stretch, which is how an undefined identifier shipped as a blank page
- [ ] Free Agents fetches every player season and filters client-side — `AFLSeason.playerSeasons` ignores `first`/`after` (`query.resolvers.go:359`). Works, but it's a season of data over the wire for a top-20 table
