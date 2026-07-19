# Current Sprint — Phase 25: FFL Scoring & Historical Import

**Sprint goal:** Make FFL scoring pluggable per season, then backfill historical FFL teams from forum data with era-correct scoring.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [x] Pluggable FFL scoring formula — per-season `Rules` value object (parameterized scoring + team structure), `ffl.season.rules_id` column, generic scoring engine; refactor-to-parity first, then define historical eras. See [ffl-scoring-rules.md](ffl-scoring-rules.md)
- [ ] FFL historical import (2006–2025) — human-in-the-loop capture (bookmarklet) + a scaffolding-first, layered import pipeline (fixtures → squads → trades → submitted teams); scores recomputed via per-season `Rules`, posted scores kept in notes for reconciliation. See [ffl-historical-import.md](ffl-historical-import.md)
  - [x] Slice 0 — nail down scoring eras (confirmed history in `rules_eras.go`; season→era tagging folded into slice 2, since seasons are created there)
  - [x] Slice 1 — capture + inspect (userscript → ingest → in-session parse; DataOps "Forum Capture" tab; validated end-to-end on real 2025 data)
  - [x] Slice 2 — season + fixtures (manual builders, no forum parsing): write-side persistence (2a); season creation with year→`rules_id` (2b); fixture builder — round-robin generator + manual rounds/fixtures — backend + DataOps "Builder" tab skeleton (2c). Superbye/byes/finals deferred (added as-we-go). **UI needs `just supergraph-compose` to expose the new mutations.**
  - [ ] Slice 3 — squads importer + closed-set player resolution (memoized per season) + review
  - [ ] Slice 4 — trades importer (`PlayerSeason` from/to-round windows)
  - [ ] Slice 5 — submitted-teams importer (post classification + authoritative-post selection; commit via `ImportRoundTeams`, evaluated scoring + posted-to-notes)
  - [ ] Slice 6 — coverage dashboard (derived) + reconciliation (per-match deltas, then spreadsheet)
- [ ] Deletion semantics fast-follow (ADR-020 step 1) — change `ffl.player_match.club_match_id` from `ON DELETE CASCADE` to `RESTRICT`, so a fixture delete that slips past the `hasTeams` guard fails loudly instead of silently destroying submitted teams. Remaining steps (`deleted_at` audit, partial unique indexes, removing the other 23 cascades) are a later slice. See [ADR-020](../ai/decisions/adr-020-deletion-semantics.md)
- [ ] Move Admin > Calculate tab to Data Ops > Calculate