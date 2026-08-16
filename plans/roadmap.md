# Roadmap

Committed phases — active and next. Completed phases 1–24 are in `plans/history.md`.

---

## Phase 26: FFL Historical Import (2006–2025)

**Goal:** Run the Phase 25 importers across the seasons, backwards from 2025, stopping wherever the source data runs out. Not pure data entry — each season may need importer changes as new formats appear.

Per season, repeated: season + clubs → squads → minor-round fixtures → teams round by round, applying trades between rounds → verify ladder and per-round scores → finals → verify → close. See [ffl-historical-import.md](ffl-historical-import.md).

- [ ] Progress table in the sprint doc, one row per season, cross-checked against committed data with SQL — no dashboard is built, the need ends with this phase
- [ ] 2025 first, proving the whole chain end to end before scaling
- [ ] Importer fixes as new squad/team/spreadsheet formats appear
- [ ] Reconciliation: per-match evaluated-vs-reference deltas, ladder at end of minor round, finals results

## Phase 27: Event Reliability (REV-1)

**Goal:** Event processing fails loudly and recoverably — remove the silent-permanent-death mode found by the 2026-07 review (`doc/review-findings.md` §6, the one `critical` finding).

- [ ] `shared/events/pg`: `Listen` reconnects with backoff on error (or the service exits fatally) — never log-and-die silently
- [ ] Re-`LISTEN` on the new connection after reconnect; log recovery at warn level
- [ ] Publish failures escalate beyond a debug/warn line (error log at minimum, visible in service output)
- [ ] Verify FFL and Search listener goroutines (`ffl/cmd/main.go`, `search/cmd/main.go`) surface listener death — no warn-and-continue
- [ ] Integration test: kill the listener connection mid-run; assert events resume after reconnect

## Phase 28: Lookup Error Handling & Type Gate

**Goal:** Finish the not-found work started in Phase 25 (`2fe2762`, `e677fc1`) and make the frontend typecheck a blocking gate. Nothing here is reachable from a routed page today — these are the same latent trap that produced the same bug three times, so they're closed together rather than one page at a time.

- [ ] Map `pgx.ErrNoRows` → `domain.ErrNotFound` in the nine remaining `FindByID` methods — AFL: `Club`, `ClubMatch`, `Player`, `PlayerMatch`; FFL: `Club`, `Player`, `PlayerMatch`, `PlayerSeason`, plus `ClubSeasonRepository.FindByClubAndSeason`. Mechanical, but they sit on command/loader paths, so run the suite alongside
- [ ] Replace the `err.Error() == "no rows in result set"` string comparison in AFL `PlayerSeasonRepository.FindLatestByPlayerID` with `errors.Is`
- [ ] Decide what `fflClub`, `fflPlayer` and `aflClub` do with an unresolvable id — they return raw parse errors and are non-nullable in the schema, so null needs a schema change. Only worth doing if something routes to a club or player page
- [ ] Clear the 7 `vue-tsc` errors and make the typecheck blocking in CI — 4 implicit-`any`, 1 unused variable, `SeasonBuilder.vue` calling `.at()` against a `lib: ES2020` target, and `TeamBuilderView.vue:507` comparing non-overlapping types (always false — possibly dead logic hiding a real bug). The typecheck was silently broken for a long stretch, which is how an undefined identifier shipped as a blank page
- [ ] Free Agents fetches every player season and filters client-side — `AFLSeason.playerSeasons` ignores `first`/`after` (`query.resolvers.go:359`). Works, but it's a season of data over the wire for a top-20 table
