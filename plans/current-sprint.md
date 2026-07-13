# Current Sprint — Phase 25: FFL Scoring & Historical Import

**Sprint goal:** Make FFL scoring pluggable per season, then backfill historical FFL teams from forum data with era-correct scoring.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [x] Pluggable FFL scoring formula — per-season `Rules` value object (parameterized scoring + team structure), `ffl.season.rules_id` column, generic scoring engine; refactor-to-parity first, then define historical eras. See [ffl-scoring-rules.md](ffl-scoring-rules.md)
- [ ] FFL historical import (2006–2025) — human-in-the-loop capture (bookmarklet) + a scaffolding-first, layered import pipeline (fixtures → squads → trades → submitted teams); scores recomputed via per-season `Rules`, posted scores kept in notes for reconciliation. See [ffl-historical-import.md](ffl-historical-import.md)