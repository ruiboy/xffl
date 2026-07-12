# Current Sprint — Phase 25: FFL Scoring & Historical Import

**Sprint goal:** Make FFL scoring pluggable per season, then backfill historical FFL teams from forum data with era-correct scoring.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [x] Pluggable FFL scoring formula — per-season `Rules` value object (parameterized scoring + team structure), `ffl.season.rules_id` column, generic scoring engine; refactor-to-parity first, then define historical eras. See [ffl-scoring-rules.md](ffl-scoring-rules.md)
- [ ] FFL historical team backfill — one-time CLI using `ForumPostParser` + `ImportRoundTeams` over historical forum data (requires pluggable scoring formula above)
