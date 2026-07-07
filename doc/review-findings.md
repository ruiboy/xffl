# Codebase Review — Findings

**Date:** 2026-07-06 · **Scope:** full repo at branch `phase-23-ux-player-intelligence` (commit `1ce978c`)

**Framing:** the repo is deliberately architecture-forward; findings judge design and execution quality on their own terms, never proportionality to project size.

**How to read a finding:** each one leads with a plain-language headline; the tags after it are `severity · good/bad/improve · classification`. For doc-vs-code mismatches, the classification says which side should change (**code drift** = fix the code, **doc should change** = fix the doc).

| Severity | Meaning |
|-----|---------|
| `critical` | Undermines a core guarantee the system claims to provide |
| `significant` | Real drift/defect; will bite as the system grows or is relied on |
| `minor` | Worth fixing; low blast radius |
| `nit` | Cosmetic / judgement call |

---

## 1. Architecture — are the rules coherent?

### Good

- **The layer rules are crisp, consistent, and genuinely enforceable.** `good` — Stated coherently across `principles.md`, ADR-005, and ADR-012. ADR-012 in particular is exemplary: domain purity, aggregate completeness, time/randomness injection, and repository responsibility each stated as enforceable rules with rationale and named negative consequences (`ai/decisions/adr-012-domain-purity-aggregate-completeness.md:74-81`).
- **Bounded contexts are applied correctly, not just claimed.** `good` — AFL and FFL share term names (`PlayerMatch`, `ClubMatch`, `Score`) with deliberately different meanings per context, documented side by side in `domain.md`; schemas are per-context; cross-context linkage is by plain integer IDs, not foreign keys (`ai/architecture/domain.md:163-165`). Textbook context-mapping (Separate Ways + ACL + published-language events).
- **ADR-013 is a model of how to reverse a decision.** `good` — History section, original premise restated, the specific evidence that invalidated it (three named traversal patterns), and an explicit consequences list touching ADR-008/015/017/018 (`adr-013:17-37`).
- **The anti-corruption layer is correct DDD.** `good` — ADR-016 + `integrations.md`: identity mapping owned by the adapter, never referenced by domain repositories, kept inside the service's own schema to preserve ADR-003's boundary.
- **The decision index is a strong human/agent interface.** `good` — `decisions.md` with status + enforceable flags, machine-readable frontmatter `rules:` on most ADRs, and an explicit source-of-truth hierarchy (`principles.md:72-80`).

### Bad / Improve

- **Two accepted ADRs give opposite instructions about calling back to the producer.** `significant · bad · docs contradict each other` — ADR-004's enforceable rule *"subscribers must not call back to the producer"* (`adr-004:9`) is contradicted by the current *designed* event flow: FFL's handlers for `AFL.PlayerMatchUpdated` / `AFL.MatchUpdated` call back to AFL over Twirp on every event (`services/ffl/internal/application/score.go:40,81`, sanctioned by `event-flow.md`). ADR-013/018 introduced sync RPC but never amended ADR-004. Either amend ADR-004 to scope the rule, or fatten the event payloads so recalculation doesn't need the callback.
- **ADR-008 still says the opposite of what was decided.** `significant · bad · doc should change` — Its frontmatter says `status: accepted` (`adr-008:2`) while `decisions.md` marks it superseded, and its closing guidance — *"Stay on path-based routing. Federation … was evaluated and rejected — see ADR-013"* (`adr-008:29`) — points at ADR-013 as the *rejection* of federation when ADR-013 *adopts* it. A reader who follows the ADR file rather than the index gets the exact opposite of reality. ADR-006 shows the correct retirement pattern.
- **"Interfaces defined where consumed" conflicts with where repository interfaces actually live.** `minor · improve` — ADR-005 states the Go idiom, but repository interfaces live in the **domain** layer (defined there, consumed by application — `services/ffl/internal/domain/club_match.go:392`), following the DDD convention instead. Clarify the rule ("repository interfaces live in domain as part of the model; all *other* ports are defined where consumed").
- **ADR-012 doesn't follow the ADR format the others use.** `minor · improve` — It lacks the standard frontmatter (status/date/scope/enforceable/rules); its enforceable rules live in prose, so any tooling reading frontmatter misses them.
- **ADR-004 contains stale facts about the event system.** `minor · bad · doc should change` — The event-flow diagram lists `FFL.FantasyScoreCalculated` (`adr-004:47`), which shipped as `FFL.PlayerMatchUpdated`; the payload-size rationale says "two event types" where there are now six (`contracts/events/events.go:5-26`).
- **The denormalised-data ADR never says how derived data stays trustworthy.** `minor · improve` — ADR-010 states "consistency maintained through domain logic on writes" but is silent on the obligations that make the pattern safe: every `drv_` field needs an idempotent rebuild path and staleness must be detectable. Both gaps materialise as concrete defects in §6. Adding those two invariants would make ADR-010 a genuinely reusable exemplar.
- **Aggregates validate on write, but persistence is column-at-a-time — so invariants can be bypassed.** `minor · improve · internal tension` — "Aggregates enforce consistency" (principles) sits awkwardly with repository interfaces that are field-level setters (`UpdateStatus`, `UpdatePosition`, `UpdateScore`, `UpdateDataStatus` — `domain/player_match.go:187-201`). The aggregate validates on `SubmitTeam`, but nothing structurally prevents writing an invariant-violating combination. See §3.

---

## 2. Domain & documentation

### Good

- **`domain.md` is an excellent ubiquitous-language document.** `good` — The position/multiplier table, bench rules as numbered hard rules mirroring `validateTeam`, the two-axis data-status → score-tier matrix, and the explicit "TM declarations are always explicit" statement all match the code precisely (star excluding hitouts: `domain.md:125` ↔ `ffl/internal/domain/player_match.go:116-120`; per-stat flooring in bye scores: `domain.md:222` ↔ `player_match.go:140-167`).
- **`event-flow.md`'s mental model is exactly what a newcomer needs.** `good` — Two inputs, everything else derived; provisional vs final tiers; and the ASCII event chain matches the handler code hop for hop.
- **The cookbook recipes are the strongest agent-facing docs in the repo.** `good` — Add-a-column with 10 ordered steps, generation-order warning, error-code convention.

### Findings

- **The docs describe FFL match storage that doesn't exist.** `significant · bad · doc should change` — `domain.md`'s Match style section says *"`clubs` stores club_season_ids … `home_club_match_id` / `away_club_match_id` are nullable"* (`domain.md:249-251`). The schema has neither column — sides are modelled by `ffl.club_match.side` (`dev/postgres/init/02_ffl_schema.sql:86`), and `match_style` exists in the DB but is absent from the Go domain (`ffl/internal/domain/match.go:18-26`), the GraphQL schema, and all logic. A parked design is presented as current fact — and per the doc's own charter, column names shouldn't be there at all.
- **The docs deny a player status ("named") that the code actively produces and stores.** `significant · bad · pick a side` — `domain.md` says twice that pre-match `named` is *not tracked* (`domain.md:90,203`), but AFL derives `"named"` whenever a player_match row exists and the match is still `no_data` (`afl/internal/domain/player_match.go:37-46`), ships it over Twirp (`afl/internal/interface/twirp/player_lookup.go:138`) and GraphQL (`afl … convert.go:126`), FFL persists it into `drv_afl_status` (`ffl/internal/application/score.go:151-154`), and FFL's GraphQL enum includes `named` (`ffl/api/graphql/query.graphqls:130-136`). Recommend **doc should change** — three layers of code already agree with each other.
- **The cookbook says Search is REST; it's GraphQL. It also misdescribes the gateway.** `minor · bad · doc should change` — Topology table at `cookbook.md:38` vs `services/search/cmd/main.go:66-74`; the gateway is described as "routes to AFL/FFL" but actually proxies `/query` to Apollo Router and `/search/query` to Search (`services/gateway/cmd/main.go:84-88`).
- **The cookbook's pagination recipe contradicts the pagination ADR.** `minor · bad · doc contradicts ADR` — The recipe puts `totalCount: Int!` on the Connection type (`cookbook.md:114-118`); ADR-014's enforceable rule puts nullable `totalCount` inside `PageInfo` (`adr-014:11`). The code follows ADR-014 (`afl/api/graphql/common.graphqls:1-5`). Fix the recipe.
- **The repo map is stale in several places.** `minor · bad · doc should change` — References `ai/architecture/service-map.md`, which doesn't exist (`repo-map.md:10`); omits `event-flow.md`, `frontend.md`, `integrations.md`, `testing.md`, `data-ops-workflow.md`; omits the search service; shows `contracts/` as events-only (no `proto/`, `gen/`); docker-compose line omits the Apollo Router (`dev/docker-compose.yml:29-40`). CLAUDE.md's `just dev-up` description has the same router omission.
- **The testing doc names e2e seed files that don't exist.** `minor · bad · doc should change` — `testing.md:232` says `dev/postgres/test-e2e/{03_afl_seed,04_ffl_seed}.sql`; the actual files are `10_afl_seed.sql` / `11_ffl_seed.sql` (confirmed by `e2e/helpers/reset-db.ts:7-10`).
- **The cookbook says schemas are shared "via symlinks"; they're actually copied.** `minor · bad · doc should change` — `cookbook.md:69` vs the `test-e2e` recipe, which copies files in and deletes them after (`justfile:116-135`).
- **The ACL mapping tables have three different naming conventions across two docs and the code.** `minor · bad` — `decisions.md` says `xref_<source>_<entity>`; ADR-016's body example is `player_source_map`; the actual tables are `afl.dataops_match_source` / `afl.dataops_player_source` (`dev/postgres/init/03_dataops.sql:4,16`); the cookbook references `<n>_<service>_integrations.sql` (actual file: `03_dataops.sql`). Pick one convention and align the rest.
- **The frontend doc describes a shared components directory that doesn't exist, and misses a whole feature.** `minor · bad · doc should change` — `frontend.md:60-77` lists `src/components/` (absent) and omits `features/data-ops/` entirely; the route table omits `/ffl/ladder` and `/afl/ladder` (`router.ts:26-28,95-97`).
- **Code comments name events that don't exist.** `nit · bad · code drift` — `team.go:44` claims SetTeam "publishes FFL.TeamSubmitted"; `team.go:254` mentions "FFL.SubsDeclared". Neither exists — both paths publish `FFL.ClubMatchUpdated`.
- **Planning docs are in genuinely good shape.** `good` — `current-sprint.md` is checked off and matches the branch's actual work; roadmap phases reference the migration strategy and carry an explicit ADR decision-gate (Phase 26).

**Newcomer verdict:** a newcomer reading `domain.md` + `event-flow.md` + `cookbook.md` could build with confidence in the FFL scoring core; the traps are the match-style fiction, the `named` status, and the topology table — all doc-side fixes.

---

## 3. Services — is the code following the rules?

### Good

- **Layer discipline is real, not aspirational.** `good` — Both services match the cookbook layout file-for-file; domain packages import only stdlib; time comes from an injected clock (`shared/clock`, `afl/cmd/main.go:146-157`), satisfying ADR-012; generated code is confined to `sqlcgen`/`generated.go`; transaction boundaries live in the application layer via `TxManager`/`WriteRepos` exactly as ADR-009 specifies (`ffl/internal/infrastructure/postgres/db.go:30-49`).
- **Service isolation holds.** `good` — No cross-service Go imports anywhere; AFL↔FFL communicate exclusively via `contracts/` events and the Twirp port (`ffl/internal/infrastructure/rpc/player_lookup.go`), matching ADR-018's port pattern precisely.
- **The ACL is honoured in practice.** `good` — External IDs live only in `dataops_*` tables accessed by dedicated repositories (`afl/internal/application/ports.go:66-86`); domain repositories never touch them.

### Drift topics, by severity

1. **The error-translation function required by ADR-009 exists but is never called.** `significant · bad · code drift` — `MapPgError` (`shared/database/errors.go:19`) is dead code. Exactly one repo method does an ad-hoc `pgx.ErrNoRows` → `domain.ErrNotFound` translation (`ffl/…/postgres/repository.go:131-132`, with `ErrNotFound` re-declared locally in `ffl/internal/domain/round.go:8`); every other `FindByID` leaks raw pgx errors up to resolvers. Either wire `MapPgError` into the repository adapters or amend ADR-009 — currently the ADR describes a fiction.

2. **Repositories return half-loaded aggregates, which ADR-012 explicitly forbids.** `significant · bad · code drift` — `ClubMatchRepository.FindByID` returns a `ClubMatch` with no `PlayerMatches`; every use case manually stitches the aggregate together (`team.go:56-71`, `score.go:161-170`). ADR-012 forbids "partially-loaded aggregates that require further data fetching" and endorses use-case-specific methods (`GetMatchWithDetails` exists for Match — the pattern is known). Relatedly, the field-level setter repositories mean team invariants are only enforced on the `SubmitTeam` path; `DeclareSubs` writes statuses/positions with no aggregate-level revalidation (`team.go:216-232`). A `SaveTeam(cm)`-style method would make the aggregate boundary real at the persistence layer.

3. **Scoring business rules are leaking out of the domain into the application layer.** `significant · improve` — The rule "an unactivated interchange bench player scores at their declared interchange position" is implemented inline in `RecalculateScore` (`score.go:136-141`) rather than in the domain (where `DeclareSubs`/`CalculateScore` live and where `domain.md` describes it). Same for "bye bench score applied at activation", split across `applyByeScoresForActivated` (`team.go:429-531`). The domain layer is where these belong and where they'd be unit-testable.

4. **Finalising a match is a chain of separate best-effort steps — a mid-sequence failure leaves inconsistent data with only a log line.** `significant · bad` — `MarkMatchStatsFinal` performs status update → result derivation → ladder recalc → event publication as separate steps outside any transaction, each failure only logged (`afl/internal/application/dataops.go:330-424`). `SetTeam`'s post-commit recalculation and publish degrade the same way (`team.go:150-185`). Combined with §6's lack of repair loops, this means permanently inconsistent derived state with no signal. At minimum, result derivation + ladder recalc for the finalised match belong in one transaction.

5. **A restore only works if the init SQL exactly matches the live schema — and nothing checks that it does.** `significant · improve` — The documented migration strategy (manual `ALTER` + keep init SQL in sync, `cookbook.md:62-75`) is coherent, but the backup is **data-only** (`pg_dump --data-only --disable-triggers`, `dev/backup/backup.sh:14`), so restoring depends on init SQL matching the schema at backup time. A forgotten init-SQL update is discovered at the next `dev-reset`/restore — the worst possible moment. Cheap fixes: also dump the schema (`pg_dump -s`) alongside each backup, and/or a recipe that diffs live schema against a fresh init-SQL database.

6. **Every schema change must be hand-applied to two independent seed datasets.** `minor · improve` — Dev seeds (6 files, `dev/postgres/seed/`) and e2e seeds (2 files, `dev/postgres/test-e2e/1{0,1}_*.sql`) are separately hand-maintained over the same schema; the cookbook recipe only mentions the dev seed (step 9). Consider generating one from the other, or documenting the dual obligation.

7. **The AFL service runs an event listener that can never receive anything.** `minor · bad · code drift` — `main.go` starts `dispatcher.Listen` with **zero subscriptions** (`afl/cmd/main.go:68-73`) — a goroutine and a pinned pool connection doing nothing.

8. **A comment promises fuzzy club-name matching that the code doesn't do — possibly a latent import bug.** `minor · bad · comment drift` — `filterByClub`'s comment says case-insensitive prefix with a contains fallback; the code is exact `==` (`afl/internal/application/dataops.go:648-658`). If FootyWire really does shorten club names, imports will silently drop a whole club's stats.

9. **The Go gateway still exists even though ADR-013 says the Apollo Router replaces it.** `minor · bad · reconcile doc or code` — ADR-013's consequence says "Apollo Router replaces `services/gateway` … CORS and `/health` move to the Router config" (`adr-013:125`); what exists is a Go proxy fronting the Router and Search (`gateway/cmd/main.go`). Keeping a stable origin is defensible — but then the ADR consequence should be revised; today the docs say the gateway shouldn't exist.

10. **The adapter documentation convention is ignored.** `nit` — `integrations.md` mandates a `doc.go` per adapter package ("the first thing a developer reads"); neither `footywire` nor `forum` has one — nor even a package comment.

---

## 4. Frontend

### Good

- **The feature-folder shape is right.** `good` — `features/{afl,ffl,data-ops}/{api,components,composables,utils,views}`, lazy-loaded route components throughout `router.ts`, GraphQL documents centralised per feature.
- **State management matches the ADR's staged intent.** `good` — Apollo cache is the only server state; the module-singleton composable pattern (`useFflState.ts:27-32`) is a clean sub-Pinia solution; `useLiveRoundBootstrap` correctly centralises the deep-link bootstrap in `App.vue`.
- **The e2e harness is the best part of the frontend estate.** `good` — See §7.

### Findings

- **The "no cross-feature imports" rule is violated everywhere, in both directions.** `significant · bad · pick doc or code` — FFL views import AFL utils/composables/queries (`ffl/views/FreeAgentsView.vue:105`, `ffl/views/RoundView.vue:71`), AFL views import FFL components/queries (`afl/views/ClubSeasonView.vue:67-70`), data-ops imports AFL wholesale (`data-ops/views/DataOpsView.vue:478-485`), and the router mounts an AFL view on an FFL-lens route (`router.ts:66-71`). Either amend `frontend.md` to define an allowed dependency direction (e.g. "ffl → afl allowed" — the AFL-lens concept implies exactly this), or create the shared `src/components`/`src/composables` home the doc already claims exists and move `Breadcrumb`, `RoundNav`, `LadderTable`, `MatchSummary`, club-logo utils, and `useXxxState` there.
- **The two most important views are unmaintainable monoliths.** `significant · improve` — `TeamBuilderView.vue` is 1,601 lines (7 queries, 3 interaction modes, drag/reorder, subs pairing, client-side bench validation, copy-previous-team); `DataOpsView.vue` is 1,022; `SquadView.vue` 590. No decomposition into per-mode components or composables, and pure logic (bench validation mirroring `domain.validateTeam`) is trapped inside a component where it can't be unit-tested. Largest single maintainability risk in the frontend.
- **The hottest pages fetch far more data than they show.** `significant · bad` — Team Builder edits **one club's** team but fetches `GET_FFL_ROUND` — every match in the round with full player lists, per-player season stats, AFL stat lines, and suggested subs for every club (`ffl/api/queries.ts:164-249`, used at `TeamBuilderView.vue:616`, re-fetched wholesale after mutations at `:1469,1520`). `SquadView` fetches `GET_FFL_SEASON_POSITIONS` — every player match of every club match of every round — to derive one squad's position history (`queries.ts:323-356`). Both need club-match-scoped queries; both multiply the server-side N+1s in §5.
- **Dead and near-duplicate GraphQL documents are accumulating.** `minor · bad` — `GET_FFL_SEASON` (82 lines, the heaviest document in the file) is **unused**; three near-identical `fflRoundByAflRound` variants coexist (`queries.ts:108-153`); there are zero fragments — the ~40-line `playerMatches` selection is pasted 6× with small accidental differences, which is exactly what fragments prevent.
- **Opening the Team Builder costs four sequential round-trips.** `minor · improve` — Club-match bootstrap → round → season → club-season (`TeamBuilderView.vue:597-643`), where the graph could serve one composed query from `fflClubMatch`.
- **There's no typecheck or lint wired into any recipe.** `nit` — The cookbook documents `npx vue-tsc --noEmit`, but `package.json` has no `typecheck`/lint script and `just test-all` doesn't run one.

### E2e test quality (coverage & brittleness)

- **Coverage maps well to the feature set.** `good` — 15 specs covering home, round, match, squad, team builder (695 lines — save, replicate, clear, subs), free agents, data-ops, navigation, plus AFL admin.
- **Some selectors will break on any styling refactor.** `minor · improve` — Structural selectors tied to Tailwind utility classes — `page.locator('div.mb-6').filter({ has: … h3 })` (`e2e/ffl-team-builder.spec.ts:14-20`) — break on a spacing change; prefer `data-testid` or landmark roles. Navigation via `page.locator('main nav').last()` (`:7`) encodes DOM order. Seed-string coupling ("The Howling Cows") is acceptable given the seeded runtime.

---

## 5. GraphQL API

### Good

- **The graph reads the way the principles say it should.** `good` — Queries start from seasons/rounds/clubs and traverse edges; the federated graph gives the frontend natural cross-context traversals (`FFLPlayerMatch.aflPlayerMatch { kicks … }`) exactly as ADR-013 intended, with `@key` stubs on the FFL side (`ffl/api/graphql/query.graphqls:156-177`) and the composed supergraph checked in.
- **Pagination follows the ADR where it exists.** `good` — `PageInfo` defined once per service with nullable `totalCount`, connections are `nodes + pageInfo` only, `first/after` accepted ahead of real pagination — matching ADR-014's staged-adoption intent.
- **The player-stats field is the best-designed part of the graph.** `good` — `AFLPlayerSeason.stats(upToRoundId, lastN, method)`: documented semantics, opaque-ID discipline stated in the docstring, reusable `AFLStatSummary` shape, and a DataLoader that batches by param-group (`afl/…/loaders.go:59-104`) — the most sophisticated loader in the repo, done right.

### Findings

- **The "everything goes through a DataLoader" rule is broken all over both services — real N+1s result.** `significant · bad · code drift vs ADR-017` — Many resolvers call single-item repo methods by ID: AFL's `Match.homeClubMatch`/`awayClubMatch` each call `GetClubMatch` + `GetClubForClubSeason` per match (`afl/…/query.resolvers.go:103-148`), `AFLClubMatch.match` calls `GetMatch` (`:38-48`), `AFLPlayerMatch.clubMatch` calls `GetClubMatch` (`:151-165`), and the season/round chains use two singles each. The loader inventory (`loaders.go:24-30`) simply lacks `ClubMatchByID`, `RoundByID`, `SeasonByID`, `ClubSeasonByID`. FFL has the same pattern (`ffl/…/query.resolvers.go:120-148,266-270`). Concrete cost: `aflRound { matches { homeClubMatch awayClubMatch } }` runs ~4 queries **per match**, and the season-wide frontend queries in §4 multiply that by rounds × matches.
- **Cross-service lookups through federation happen one row at a time.** `significant · bad · code drift vs ADR-013` — ADR-013 says "entity resolvers batch by IDs using the existing dataloadgen pattern"; the implementation is per-representation `Find*ByID` calls (`afl/…/entity.resolvers.go:26-40`), so an FFL club-match page resolving ~30 `aflPlayerMatch` references issues ~30 sequential single-row queries inside one `_entities` call (gqlgen's batched entity-resolver mode is unused). Exactly the cross-service N+1 that ADR-017's "Upcoming pressure" section predicted.
- **A query root exists that the principles explicitly forbid.** `significant · bad · pick doc or code` — `principles.md:34` names **ClubMatch** as an internal join entity that "must not be a query root", yet `fflClubMatch(id:)` is a root (`ffl/api/graphql/query.graphqls:15`) and Team Builder routes are keyed on club-match IDs. In practice FFL ClubMatch has grown into a page-level aggregate ("a club's team for a round"). Recommend **doc should change** — but restate the principle so it still excludes true join noise.
- **The two subgraphs model the same kinds of values differently.** `minor · bad` — FFL models statuses as enums (`FFLPlayerMatchStatus`, `FFLAFLPlayerMatchStatus`); AFL exposes `status: String!`, `dataStatus: String!`, `result: String` as bare strings (`afl/api/graphql/query.graphqls:47-56,142`). Same domain, one graph, two conventions — enums should win.
- **Whether a missing ID errors or returns null is arbitrary per field.** `minor · bad` — `aflSeason(id): AFLSeason!` (error on missing) vs `aflRound(id): AFLRound` (null); `fflClub!` vs `fflClubSeason` — no discernible rule. Pick "nullable + null on not-found" for lookups and apply it.
- **Some entity resolvers run a query just to throw the result away.** `minor · improve` — `FindAFLPlayerSeasonByID` fetches the row, discards it, and returns an ID-only stub (`entity.resolvers.go:43-54`) — a query per representation purely to 404. Batched entity resolvers make this cost disappear.
- **A few unbounded lists have no pagination.** `minor · improve` — Mostly defensible under ADR-014's ~50-item rule; the exceptions are `fflPlayers` (grows every season, unbounded) and `AFLClubSeason.playerSeasons`.
- **The schema documents a filter as "Not supported".** `nit` — `FFLPlayerSeasonFilter.active` (`query.graphqls:117-120`) — remove until real.
- **Frontend query documents** (over/under-fetch, fragments): covered in §4 — over-fetch significant, zero fragment reuse, one dead document.

---

## 6. Events & data consistency

### Good

- **The designed chain is conceptually strong.** `good` — `AllAFLStatusesFinal` is an elegant finality inference — FFL derives AFL finality from its own data with no cross-service call at check time (`event-flow.md:57-62`; SQL-backed at `ffl/…/player_match.go:199`) — and `dnp` inference via the both-squads status map on `final` is correctly specified and implemented (`afl/…/dataops.go:397-410`).
- **Ladder recalculation is rebuild-from-scratch idempotent, and repair mutations exist.** `good` — "Entire season recalculated from scratch — simpler and drift-free"; `recalculateFFLLadder`, `recalculateFFLClubMatchScore`, `recalculateAFLLadder`.
- **Derived scores are kept consistent inside the writing transaction.** `good` — Every synchronous write path re-derives `drv_score` in the same transaction as the player-match write (`SetTeam`, `CalculateFantasyScore`, `RecalculateScore`, `ResolveAFLPlayerMatch`) — ADR-010 honoured on the happy path.
- **The event chain has dedicated integration tests.** `good` — `ffl/…/event_integration_test.go`, 842 lines — rare and valuable.

### Findings

- **A single lost event permanently stalls match finalization, and nothing will ever notice.** `critical · bad` — Four compounding facts:
  1. Publishes happen post-commit on a separate connection and every failure is warn-and-continue (`team.go:182-184`, `dataops.go:302-304`, etc.) — a crash or error between commit and publish loses the event.
  2. There is no outbox/replay (ADR-004 consciously accepts this, deferring the outbox to its scale path).
  3. If `Listen` errors (DB restart, connection drop), the goroutine logs "listener stopped" and **never restarts** — the service keeps serving traffic with event processing permanently dead (`ffl/cmd/main.go:97-101`, `search/cmd/main.go:58-62`, `shared/events/pg/dispatcher.go:70-88`).
  4. The user-visible finality pipeline — `ClubMatchScoreFinalized` → `MatchScoreFinalized` → `drv_result` + ladder — is driven **only** by this chain; `markFFLTeamFinal` returns success after a status write and a publish (`ffl/…/dataops.go:222-248`), and everything after is fire-and-forget.

  ADR-004's "lost messages are tolerable because everything can be recalculated" only holds if recalculation paths exist — and `ffl.match.drv_result` has **no** manual rebuild (`RecalculateFflLadder` never calls `UpdateResult`; only `ProcessFflMatchScoreFinalized` does, `reactions.go:255`). Minimum fixes, in leverage order: (a) make `Listen` reconnect with backoff or crash the process — silent degradation is the worst outcome; (b) extend a repair mutation to re-derive `drv_result` and re-run finalization inference for a round; (c) longer-term, the ADR-004 outbox.

- **Un-finalising an AFL match leaves FFL believing it's still final.** `significant · bad` — `MarkMatchStatsFinal(matchID, final=false)` updates AFL's status and returns early — **no event** (`afl/…/dataops.go:347-349`). FFL keeps `drv_afl_status = played/dnp` for every affected player, so `AllAFLStatusesFinal` stays true and a subsequent FFL-side finalization locks scores against AFL data that is officially no longer final. The reversal path must emit `AFL.MatchUpdated(partial)` with a re-derived status map (the FFL side gets this right — `markFFLTeamSubmitted` publishes on reversal, `ffl/…/dataops.go:193-218`).

- **There is no way to detect that derived data has gone stale.** `significant · bad` — Nothing compares stored `drv_score` against a recomputation, or detects club matches that satisfy both-axes finality but never emitted `ClubMatchScoreFinalized`. Every failure in this section is therefore *invisible* until a human notices a wrong ladder. The data-ops workflow already plans "Step 6 — Score reconciliation" (`data-ops-workflow.md:66-73`) — building it as a systematic drv-vs-derived diff closes this hole and satisfies the ADR-010 invariant proposed in §1.

- **One AFL match finalising triggers a storm of calls back to the AFL service.** `significant · bad · see §1` — Every `AFL.MatchUpdated` triggers, per FFL club match in the round, `RecalculateScore` → up to 2 Twirp calls back to AFL plus per-player queries (`reactions.go:115-139`); one `AFL.MatchUpdated(final)` fans out to ~8 club matches × (Twirp + N queries), serially, on the single listener connection. Contradicts ADR-004's no-callback rule and amplifies load on the producer at exactly its busiest moment. Fat events (stats snapshots in the payload) or a batched lookup fixes both.

- **One slow event handler stalls all event processing for the whole service.** `minor · bad` — Handlers run inline and sequentially in the `Listen` loop with the root context and no per-event timeout (`dispatcher.go:101-105`); a hung Twirp call blocks every subsequent event for the process.

- **Tests exercise stricter event semantics than production has.** `minor · bad` — The in-memory dispatcher propagates the first handler error to the publisher (`memory/dispatcher.go:32-37`); the PG dispatcher swallows handler errors (`pg/dispatcher.go:101-105`). A path to "passes in test, silently drops in prod".

- **Finalized events can re-fire; it's benign today, but only by accident.** `minor` — `ClubMatchScoreFinalized` re-fires on every reprocessing of an already-final club match (`reactions.go:130-137`); downstream is idempotent so this is harmless — worth a comment so it stays deliberate.

- **Nothing guards the 8KB NOTIFY payload ceiling.** `nit` — Current payloads (~1–3KB for full-squad maps) have headroom, but nothing checks or logs payload size at publish; an oversized snapshot would fail silently into a warn log.

---

## 7. Testing strategy adherence

### Good

- **Tier discipline is largely real.** `good` — Every `*_integration_test.go` in AFL/FFL carries `//go:build integration`; the `TestMain`-per-package + `testutil.StartPostgres` + truncate-and-seed isolation pattern is implemented exactly as `testing.md` specifies; external HTTP is mocked with `httptest` everywhere (no real network calls found).
- **Integration coverage is substantial and aimed at the risky seams.** `good` — AFL GraphQL (1,727 lines), FFL GraphQL (1,625), the event chain (842), data-ops import (444), suggested substitutions (447), subs declaration (386), byes (238), Twirp server (374). Testing the event cascade end-to-end at the integration tier is the strategy's standout.
- **Domain unit tests follow the documented conventions.** `good` — Table-driven, sentence-style case names, want-first asserts; `ffl/internal/domain/club_match_test.go` (19KB) thoroughly covers `validateTeam`, `DeclareSubs`, `Score`, and `SuggestedSubstitutions` — the highest-value pure logic in the system.
- **The e2e isolation model matches its documentation.** `good` — Auto-reset fixture, `workers: 1` with the rationale in a config comment as promised, fixed `CLOCK_OVERRIDE`, fully isolated stack on separate ports.

### Findings

- **Two test files break the promise that plain `go test ./...` is always safe to run.** `significant · bad · code drift` —
  1. `services/search/internal/infrastructure/typesense/repository_test.go` starts a **testcontainers** Typesense with no build tag — `go test ./...` in the search service requires Docker, violating `testing.md`'s explicit rule.
  2. `shared/events/pg/dispatcher_test.go` connects to the **live dev database** (`localhost:5432/xffl`, `dispatcher_test.go:17`), untagged — it fails without the dev stack and touches real shared infrastructure, violating both the tag rule and the hermeticity rule.
- **The search service and shared packages are never tested by any recipe.** `significant · bad` — `just test-all` runs AFL, FFL, and e2e only (`justfile:138-141`); the two violations above (and any future regressions in those modules) are invisible in the normal workflow.
- **Coverage gaps line up with the layering drift.** `minor · improve` — AFL's `CalculateLadder` has no unit test (the FFL twin does — `ffl/…/ladder_test.go`); the both-axes finalization inference and status-map application live in `application/reactions.go` with repository dependencies, so they're only covered via heavyweight integration tests — a direct consequence of §3 finding 3. There's no test for listener failure/reconnect behaviour (because there is none to test — §6).
- **The docs claim strict TDD; the history can't support that claim.** `minor · improve` — Principles list TDD as non-negotiable ("write failing tests first"), which is unverifiable from the history (tests and implementation land together in feature commits). The honest, still-strong claim the repo *does* satisfy is "no behaviour ships untested at the appropriate tier". Reword, or accept it as aspirational process guidance for agents.
- **Pure frontend logic is only tested through the slowest possible tier.** `minor · improve` — `position.ts`, `scoring.ts`, `heatmap.ts`, and the bench-validation logic trapped in `TeamBuilderView.vue` have zero direct tests — consistent with the documented "frontend = e2e only" strategy, but these are exactly the cheap-to-unit-test functions where e2e is the wrong tier. A small vitest setup closes this without changing the strategy's spirit.
- **E2e selector brittleness** — see §4. `nit`

---

## Cross-lens prioritised actions

1. **Make event processing fail loudly and recoverably** *(§6 critical)* — reconnect-with-backoff (or fatal exit) when `Listen` dies in FFL/search; log-and-alert on publish failure. This single change removes the silent-permanent-death mode. (`shared/events/pg/dispatcher.go`, both `cmd/main.go`s.)
2. **Give every derived field a rebuild path and a staleness check** *(§6 + ADR-010)* — extend repair mutations to re-derive `ffl.match.drv_result` and re-run finalization inference for a round; build data-ops Step 6 as a drv-vs-recomputed reconciliation diff. Then amend ADR-010 to state both as invariants of the pattern.
3. **Emit `AFL.MatchUpdated(partial)` on final→partial reversal** *(§6)* — with a re-derived status map, mirroring what `markFFLTeamSubmitted` already does on the FFL side (`afl/internal/application/dataops.go:347`).
4. **Finish ADR-017 and batch the federation entity resolvers** *(§5)* — add `ClubMatchByID`/`RoundByID`/`SeasonByID`/`ClubSeasonByID` loaders in AFL (mirror in FFL), convert the direct `Get*` resolver calls, and switch gqlgen federation to batched entity resolvers. This collapses the worst N+1s, including the cross-service one.
5. **Reconcile the contradicting ADRs** *(§1)* — mark ADR-008 superseded in its own frontmatter and correct its "federation rejected" sentence; amend ADR-004's no-callback rule (or move to fat events) and refresh its stale event names/counts; align ADR-016/decisions.md/cookbook on one ACL table-naming convention.
6. **Wire `MapPgError` (or delete it and amend ADR-009)** *(§3)* — one adapter-level change per service; kills the raw-pgx-error leak and the duplicated `ErrNotFound`.
7. **Fix the test-tier violations and recipe gaps** *(§7)* — tag the Typesense container test `integration`; rewrite the pg dispatcher test on testcontainers with a tag; add search + shared to `just test-all`.
8. **Documentation truth pass** *(§2, one sitting)* — `domain.md` (match-style storage fiction, `named` status), `cookbook.md` (Search=GraphQL, gateway role, symlink→copy, pagination recipe `totalCount`), `testing.md` (seed file names), `repo-map.md` (service-map.md, router, search, contracts), `frontend.md` (cross-feature rule, layout, data-ops).
9. **Frontend structural pass** *(§4)* — decide the cross-feature import rule and create the shared `src/components`/`composables` home; split `TeamBuilderView.vue` into mode components + composables; introduce fragments for the repeated player-match selection; scope Team Builder and Squad views to club-match-level queries instead of round/season-wide documents.
10. **Backup/restore parity guard** *(§3)* — dump schema alongside data backups (or add a recipe diffing `pg_dump -s` against a fresh init-SQL database) so manual-migration drift is caught before a restore depends on it.
