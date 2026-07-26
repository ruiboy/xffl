# Current Sprint — Phase 25: FFL Scoring & Historical Import

**Sprint goal:** Make FFL scoring pluggable per season, and build the tooling to backfill historical FFL teams from forum data with era-correct scoring. Running that import across 2006–2025 is Phase 26.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [x] Pluggable FFL scoring formula — per-season `Rules` value object (parameterized scoring + team structure), `ffl.season.rules_id` column, generic scoring engine; refactor-to-parity first, then define historical eras. See [ffl-scoring-rules.md](ffl-scoring-rules.md)
- [ ] FFL historical import **tooling** — the importers, not the import. Running them across 2006–2025 is Phase 26. Scores recomputed via per-season `Rules`; every score found (forum-posted or spreadsheet) is a reference kept in notes, never `drv_score`. See [ffl-historical-import.md](ffl-historical-import.md)
  - [x] Slice 0 — nail down scoring eras (confirmed history in `rules_eras.go`; season→era tagging folded into slice 2, since seasons are created there)
  - [x] Slice 1 — capture + inspect (userscript → ingest → in-session parse; DataOps "Forum Capture" tab; validated end-to-end on real 2025 data)
  - [x] Slice 2 — season + fixtures (manual builders, no forum parsing): write-side persistence (2a); season creation with year→`rules_id` (2b); fixture builder — round-robin generator + manual rounds/fixtures (2c). Byes, superbye and finals added as-we-go. **UI needs `just supergraph-compose` to expose the new mutations.**
  - [x] Slice 3 — squad importer: squads-thread parser + bulk resolution-review UI, plus a raw-text passthrough in the capture preview (`ForumCaptureBuffer.Ingest` currently discards raw HTML, so an unparseable thread shows nothing). Trades stay manual on the existing squad page
    - Backend done + tested: parser (`forum/squads.go`); `ParseSquadThread`/`ParseSquadThreadForSeason`/`ImportSquad` (`dataops/squads.go`); raw-text passthrough (`PreviewedPost.Text`); candidate sourcing via new AFL `ListSeasonPlayers` twirp endpoint + FFL `LookupSeasonPlayers`.
    - GraphQL + UI done: `parseFFLSquadThread`/`importFFLSquad` mutations (`dataops.graphqls` + resolvers, supergraph recomposed); "Import Squads" DataOps tab — paste thread → per-club review with auto club-assignment, cost, confidence, per-member "Fix" via `SquadMemberLinkModal`, per-squad import. Resolver parse-path test in `squad_import_test.go`.
  - [ ] Slice 4 — spreadsheet fixture importer (pasted season sheet → rounds, fixtures, reference club scores to `notes`). Fixtures are upstream of teams in the per-season workflow (step 3 before step 4), so it lands before the submitted-teams importer
    - Backend done + tested (unit + integration): parser (`spreadsheet/fixtures.go`); `ImportFixtures` (`dataops/fixtures_import.go`) builds rounds/fixtures + inferred byes via `SaveFixtures` and writes reference scores to `club_match.notes`. **Remains: GraphQL + review UI; superbye/finals nuance left to the manual builder.**
  - [ ] Slice 5 — submitted-teams importer, H&A and finals alike: season-scoped author→club_season registry, post classification, authoritative-post selection, commit via `ImportRoundTeams` **+ per-match evaluated-vs-reference delta view** (ships here — deltas are the only proof a season's `rules_id` is right)
    - Not started. **Open design question:** a post's club is not reliably the author's — anyone can post another club's team (e.g. THC's was posted on their behalf). Attribution must come from the post itself; if absent, stop and ask, or infer from the named players' club_season in the prior round. The earlier author→club_season registry idea is dropped.
- [ ] Deletion semantics fast-follow (ADR-020 step 1) — change `ffl.player_match.club_match_id` from `ON DELETE CASCADE` to `RESTRICT`, so a fixture delete that slips past the `hasTeams` guard fails loudly instead of silently destroying submitted teams. Remaining steps (`deleted_at` audit, partial unique indexes, removing the other 23 cascades) are a later slice. See [ADR-020](../ai/decisions/adr-020-deletion-semantics.md)
- [x] Move Admin > Calculate tab to Data Ops > Calculate