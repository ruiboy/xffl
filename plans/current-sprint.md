# Current Sprint — Phase 25: FFL Scoring & Historical Import

**Sprint goal:** Make FFL scoring pluggable per season, then backfill historical FFL teams from forum data with era-correct scoring.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [ ] Pluggable FFL scoring formula — strategy pattern keyed by season; `ScoringStrategy` interface + concrete implementations covering known formula variants; `ffl.season.scoring_strategy` column; wire into score calculation use case (deferred from Phase 20)
- [ ] FFL historical team backfill — one-time CLI using `ForumPostParser` + `ImportRoundTeams` over historical forum data (requires pluggable scoring formula above)
