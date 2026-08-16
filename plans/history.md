# History — Completed Phases

Completed phases from the xffl rebuild. See `plans/roadmap.md` for active and upcoming phases.

---

## Phase 1: Foundation ✅

**Goal:** Dev environment + shared packages + contracts

- [x] `dev/docker-compose.yml` — PostgreSQL, Zinc
- [x] `justfile` — recipes: `dev-up`, `dev-down`, `dev-reset`, `dev-seed`
- [~] Migration tooling — keeping raw SQL init scripts for now
- [x] `shared/database/` — DB connection helper
- [x] `shared/events/` — event dispatcher interface + PG LISTEN/NOTIFY implementation
- [x] `shared/events/memory/` — in-memory dispatcher for testing
- [x] `contracts/events/` — shared event type definitions

## Phase 2: AFL Service ✅

**Goal:** Complete AFL service with sqlc

- [x] Domain layer — entities + repository interfaces
- [x] Migrate infrastructure to sqlc (ADR-009 compliance)
- [x] Application layer — mutations with `DB.WithTx` transaction support
- [x] Interface layer — GraphQL schema + resolvers + HTTP server
- [x] Tests — unit (domain) + integration (GraphQL with real DB)

## Phase 3: UX Scaffold ✅

**Goal:** Gateway + Vue 3 app scaffold + first AFL view with edit capability

- [x] ADR-008 — gateway as simple reverse proxy
- [x] ADR-011 — frontend stack (Vue 3, Apollo, Tailwind, PrimeVue unstyled)
- [x] Gateway — GraphQL proxy routing to AFL, CORS, health checks
- [x] Vue 3 project setup — TypeScript, Vite, Apollo Client (pointing at gateway :8090), router, Tailwind
- [x] AFL Match view — match result with player stats, inline editing, mutations via Apollo
- [x] Playwright e2e tests for match view (read + edit, 6 tests)

## Phase 4: AFL Frontend ✅

**Goal:** Remaining AFL views + UX polish

- [x] AI plans & prompt improvements
- [x] AFL frontend page discovery — interview, document page inventory, confirm scope
- [x] Add `aflLatestRound` backend query + `season` field on `AFLRound`
- [x] Build pages — Home (ladder + matches + round nav), Round (matches + top players + round nav), Match (read-only with club logos), Admin Match (editable player stats)
- [x] Playwright e2e tests (16 tests across home, round, match, admin-match)
- [x] UX polish — light theme, AFL club logos (18 teams), semantic Tailwind v4 `@theme` tokens, light/dark theme switcher
- [~] PrimeVue unstyled — deferred until complex interactive components are needed (Phase 5/6)

## Phase 5: FFL Service ✅

**Goal:** FFL service as standalone CRUD with position-based fantasy scoring

- [x] Domain layer — League, Season, Round, Match, ClubSeason, ClubMatch, Player, PlayerSeason, PlayerMatch entities; position-based scoring (goals/kicks/handballs/marks/tackles/hitouts/star); bench + interchange substitution logic; repository interfaces
- [x] Application layer — ManagePlayers (CRUD), QueryLadder, CalculateFantasyScore use cases
- [x] Infrastructure layer — sqlc queries, DB repositories, transaction manager
- [x] Interface layer — GraphQL schema + resolvers
- [x] Add FFL routing to gateway
- [x] Tests — unit (scoring by position, percentage, substitution) + integration (GraphQL with real DB)

## Phase 6: FFL Frontend ✅

**Goal:** FFL views added to existing frontend

- [x] FFL Players view
- [x] Playwright tests

## Phase 7: Data Model Refinements ✅

**Goal:** Clean up AFL/FFL data models and propagate changes through the stack

- [x] Remove `afl.player.club_id` and `ffl.player.club_id` — players exist independently; they attach to clubs for a season via `player_season → club_season`
- [x] Add `from_round_id`/`to_round_id` to `afl.player_season` for mid-season transfers
- [x] Add `status` (named/played/dnp) to `afl.player_match`
- [x] Rename `ffl.player.name → drv_name`, `ffl.player_match.score → drv_score`; domain entities keep clean names
- [x] Add unique constraint on `ffl.club_match(club_season_id, match_id)`
- [x] Drop `AFLClub.players` GraphQL field
- [x] Update domain entities, GraphQL schemas/resolvers, frontend, seed data, and tests
- [x] All 43 Playwright e2e tests green; full Go test suite green

## Phase 8: FFL UX Refinements ✅

**Goal:** Iterative FFL frontend improvements, driven by user requests each session

- [x] Rename Roster → Squad throughout
- [x] FFL pages routed under `/ffl`; nav + home link updated; `/` redirects to `/ffl`
- [x] Architecture: no graph federation; CQRS read/write split decided (ADR-013); gateway routing clarified (ADR-008)
- [x] Fix wasteful squad query — `fflClubSeason(seasonId, clubId)` resolver; connection pagination
- [x] Fix fragile Apollo routing link — explicit operation-name map replaces regex
- [x] Global FFL club state (`useFflState`) + unified nav with club selector
- [x] Home/round page layout: circle round selector (filled/ring/ladder icon), inline headings, no matches on home
- [x] FFL eagle logo in nav (hover scales 3×)
- [x] Settings cog dropdown with dark mode toggle (cookie-persisted)
- [x] Squad page: club name heading, search panel alongside player list, Manage/Done pattern
- [x] Team Builder: club name heading, Manage/Done pattern (Done saves team)
- [x] `FFLClubSeason.season` field added to GraphQL schema and resolver

## Phase 9: FFL Team Composition Rules ✅

**Goal:** Define and enforce rules for how an FFL team is structured each round

- [x] Clarify team composition rules (positions, required structure, bench/interchange constraints)
- [x] Domain logic + validation (`ValidateTeam`, multi-starter `Score()` fix)
- [x] Enforce validation in `SetTeam` command + GraphQL mutation
- [x] Team Builder UI rebuilt with structured position layout + bench/interchange management
- [x] Domain unit tests + GraphQL integration tests

## Phase 10: Test Stabilisation ✅

**Goal:** Refactor tests to be accurate, extensible, and grounded in minimal seed data

- [x] Migrate AFL + FFL integration tests from dev Postgres to testcontainers (hermetic, no shared state)
- [x] `TestMain` + shared container pool pattern; per-test `t.Cleanup` truncates
- [x] Refactor all tests to testify (`require`/`assert`) with sentence-style `t.Run` names
- [x] Add `ai/architecture/testing.md` conventions doc + `/write-tests` skill
- [x] Delete mock-based `commands_test.go` — coverage consolidated into integration tests

## Phase 11: FFL Event Integration ✅

**Goal:** Wire up cross-service event flow between AFL and FFL

- [x] Contract extended: `RoundID` added to `PlayerMatchUpdatedPayload`
- [x] AFL publishes `AFL.PlayerMatchUpdated` after stat updates (PG LISTEN/NOTIFY)
- [x] FFL round correlation: `afl_round_id` column + join query for player_match lookup
- [x] FFL subscribes to `AFL.PlayerMatchUpdated` → auto-calculates fantasy scores
- [x] FFL publishes `FFL.FantasyScoreCalculated`
- [x] Tests — integration (event flow end-to-end, unknown player, multiple clubs)

## Phase 12: Live Round ✅

**Goal:** Compute and expose a contextually relevant "live round" across AFL and FFL, drive round nav defaults and indicators from it, and make the whole thing testable without real-time dependency.

- [x] Injectable `Clock` interface (shared); `CLOCK_OVERRIDE` env var for e2e
- [x] AFL `LiveRound` use case — `FindNeighbours` DB query + Adelaide midnight boundary; nullable (nil before season starts); no Open/Closed status — single ring style
- [x] AFL `aflLiveRound` GraphQL query (nullable)
- [x] FFL maps AFL→FFL round client-side via `afl_round_id`; no FFL service-side live round query needed
- [x] E2e seed data with fixed `start_dt` values; `CLOCK_OVERRIDE` wired into Playwright config
- [x] Frontend: `useAflState` + refactored `useFflState` — JSON cookies `xffl_afl` / `xffl_ffl` with `{ seasonId, roundId, startDate }`
- [x] RoundNav: `liveRoundId` from cookie; single `ring-active` indicator (no closed/open distinction)

## Phase 13: Search Service ✅

**Goal:** Event-driven search indexing

- [x] Domain layer — SearchDocument, SearchQuery, SearchResult
- [x] Application layer — Search, IndexDocument use cases; event handlers for indexing
- [x] Infrastructure layer — Typesense client + repository (testcontainers integration tests)
- [x] Interface layer — GraphQL (`search` query, playground, health)
- [x] Add search passthrough to gateway (`/search/query`)
- [x] Tests — unit (payload→document transformation) + integration (Typesense round-trip, filtering, upsert)

## Phase 14: Historical AFL Data — afltables (2024–present) ✅

**Goal:** Load real AFL player match stats into the domain for 2024 onward using afltables.com. Establishes the canonical player roster.

- [x] Generate `dev/postgres/seed/03_afl_historical.sql` directly from afltables CSV files (810 players, 21942 player_match rows)
- [x] Seed FFL Ruiboys squad (30 players) in `04_ffl_players.sql`
- [x] Load into dev DB; verify player/match counts and spot-check stats

## Phase 15: Database Backup ✅

**Goal:** Persist DB state durably outside the dev lifecycle.

- [x] `just backup-db` — `pg_dump --data-only | gzip` → timestamped local file; uploads via rclone if `BACKUP_REMOTE` set
- [x] `just restore-db` — reset dev DB, restore from backup, verify row counts
- [x] Restore verified end-to-end (810 players, 21942 player_match, 30 ffl players)

## Phase 16: 2026 FFL Data Import ✅

**Goal:** Seed real 2026 FFL data for rounds 1–5 (R6 squads only, no scores yet).

- [x] Identify data source — Tapatalk forum posts (manual copy-paste)
- [x] Build `dev/import/ffl/parse_forum.py` — parses all 4 team formats → `*_teams.csv` + `*_scores.csv`
- [x] Parse and validate R1–R6 (88 player rows/round); R6 squads parsed (no scores)
- [x] `resolve_squads.py` + `import_round_teams.py` — stopgap Python importers; seed SQL 04–06 generated with name-based subqueries (no hardcoded IDs)
- [x] 120 FFL players + 4 mid-season trades seeded; R1–R5 player_match + scores inserted
- [x] `dev-seed` runs all 6 seed files end-to-end (idempotent)
- [x] Verify ladder standings and scores post-import

## Phase 17: UX Improvements ✅

**Goal:** Richer, faster frontend — more detailed stats, new player and team pages, and meaningful performance improvements as data volume grows.

- [x] Performance: break up monolithic GraphQL queries — FFL (done); AFL RoundView + MatchView (done); N+1 batch fix for AFL/FFL PlayerMatches + Ladder resolvers (done); query count logging via pgx tracer (done)
- [x] DataLoader pattern (ADR-017) — per-request `Loaders` struct injected via context; `vikstrous/dataloadgen`
- [x] FFL Team Builder UX — player scores alongside names, position group totals, grand total in team summary bar, status badges, scoring formula for multiplier positions (`utils/scoring.ts`)
- [x] AFL MatchView — Manage button icon
- [x] E2e test isolation — per-test DB reset via auto worker fixture (`fixtures.ts`); `TRUNCATE … RESTART IDENTITY CASCADE` for stable IDs; `helpers/reset-db.ts` via `docker exec`; `workers: 1`; restored `ffl-team-builder.spec.ts`
- [x] E2e docs — `ai/architecture/testing.md` Playwright section; cookbook recipe for adding new specs

## Phase 18: Data Management — Import Infrastructure, Part I ✅

**Goal:** Build recurring data flows for team submissions, AFL stats, score reconciliation, historical backfill, and season setup. All Go; ports-and-adapters throughout; Twirp for cross-service calls.

- [x] ADR — Twirp for cross-service communication
- [x] FFL Round team submission
- [x] AFL stats import

## Phase 19: Graph Federation ✅

**Goal:** Adopt Apollo Federation so the frontend can traverse cross-service entity relationships in a single query. Replace the path-based gateway with Apollo Router. Establish `AFLPlayerSeason` as a first-class graph type.

- [x] Apollo Router — replace `services/gateway` with Apollo Router in Docker Compose; configure supergraph from both subgraphs
- [x] AFL subgraph — add `AFLPlayerSeason` type + entity resolver; add `@key` to `AFLPlayerMatch`; mount federation-compatible handler
- [x] FFL subgraph — add `aflPlayerSeason` field on `FFLPlayerSeason`; add `aflPlayerMatch` field on `FFLPlayerMatch`; reference resolvers; mount federation-compatible handler
- [x] Frontend — single Apollo client endpoint; remove operation-name routing link
- [x] Tests + e2e verification

## Phase 20: Data Management — Import Infrastructure, Part II ✅

- [x] FFL in-season player trades — squad management with AFL-backed player search
- [x] AFL stats import — FootyWire scraper, player source mapping, match status tracking
- [x] Score and ladder calculation — AFL→FFL event chain, finalization flow, provisional/final tiers
- [x] Schema health: replace circular match↔club_match FKs with role column; enforce AFL FK integrity
- *(deferred → Phase 23)* Score reconciliation — submitted vs. calculated diff
- *(deferred → Phase 25)* Pluggable FFL scoring formula — strategy pattern keyed per season
- [x] Close out: drop `ffl.player.drv_name`, retire `parse_forum.py`, move stats import status to dataops table; add `notes` to `ffl.club_match` and `ffl.player_match`

## Phase 21: UX — Navigation ✅

**Goal:** Consistent, round-aware navigation across AFL and FFL — header links land on the live round, cross-domain switching is a first-class affordance, and DataOps is always reachable.

- [x] NAV-1: Update header AFL/FFL links to navigate to live round (`/afl/rounds/:liveRoundId`, `/ffl/rounds/:liveRoundId`)
- [x] NAV-2: Cross-domain round-aware navigation — header AFL/FFL/DataOps links follow a session-scoped "selected round" (defaults to live round, sticks to whichever round you're viewing, syncs the corresponding round across domains); replaces the originally-planned cross-domain pill, which was built then superseded by this approach
- [x] NAV-3: Add Ladder pill as first item in RoundNav (ladder icon + hover tooltip)
- [x] NAV-4: Add DataOps icon link to header right-side nav (always visible; links to live round)
- [x] NAV-5: Replace DataOps round dropdown with RoundNav component
- [x] PAGE-7: Enhance current-round pill with pulsing indicator when `start_dt` = today
- [x] Playwright tests for new navigation elements

## Phase 22: Real Data Load & Seed Cleanup ✅

**Goal:** Replace synthetic seed data with real 2026 AFL and FFL data, and slim the dev seed to a lean scaffold.

- [x] Import all 2026 AFL rounds played to date (teams, players, match stats) via existing import tooling
- [x] Import all 2026 FFL round teams to date
- [x] Verify ladder calculation produces correct standings
- [x] Smoke-test team submission and substitution flows against real data; capture edge cases
- [x] Reduce `dev/seed` to a single representative round (enough for demo and future dev)
- [x] Confirm e2e tests pass against slim seed

## Phase 23: UX — Player Intelligence ✅

**Goal:** Richer player and club context for FFL decision-making. Candidate areas from `plans/ideas.md`: player season pages, free agents, club pages, Team Builder analytics, notes in the match view, score reconciliation. Scope confirmed at sprint start.

- [x] PAGE-1: AFL player season view (`/ffl/afl/player-seasons/:id`) — MEDIAN backend, status pills, season stats table, row-level match links, header breadcrumb links, stat label fixes
- [x] PAGE-3: AFL club season view (`/ffl/afl/club-seasons/:id`) — all players at an AFL club with FFL ownership, stats, and ★ avg; entry from AFL ladder club names
- [x] Club Match page (`/ffl/club-matches/:id`) — read-only view of a club's team and scores; linked from FFL match view; Team Builder link shown when club is selected
- [x] Frontend architecture: formalised three-namespace page hierarchy (FFL / AFL / AFL Lens) in `frontend.md`; consistent page naming and route conventions across `router.ts`, `frontend.md`, and `ideas.md`
- [x] Cross-page linking UX pass — AFL round Top Players (logo + match link), AFL match DataOps footer removed + stat headers (K H D M R T G B Pts), AFL Lens club season inline player/club links, FFL round top scorer match links, FFL match header club season links, "Improve your score" pill moved above grid for alignment
- [x] Team Builder: drag-and-drop with popup-menu fallback, per-player stat summaries with heatmaps and form/season trend arrows, projected team total, and Best Team optimiser (Hungarian assignment)
- [x] Shared stat intelligence: global Last N/Season source toggle, `trendDir` thresholds (constants), sortable stat columns, richer player hover card (per-round heat-mapped lines); vitest unit tests (ADR-019)

## Phase 24: Data Management — Data Setup & Historical Import ✅

**Goal:** Season setup tooling and one-time historical backfill — the operations needed once per season (or once ever) rather than every round.

- [x] AFL historical data import — afltables scrape + import CLIs; 1998–2023 loaded, 2024/2025 finalised + derived (scores, rushed behinds, ladders). Made navigable in the webapp (season picker, season-in-URL, ladder home) and finals typed via `round_type`.

## Phase 25: FFL Scoring & Historical Import Setup ✅

**Goal:** Make FFL scoring pluggable per season, then backfill historical FFL teams from forum data with era-correct scoring.

- [x] Pluggable FFL scoring formula — per-season `Rules` value object with `Scoring` and `Composition` facets (`ffl.season.rules_id`) + generic scoring engine; refactor-to-parity first, then historical eras defined in `rules_eras.go`. Supersedes the `ScoringStrategy` interface / `scoring_strategy` column originally scoped here — parameterised rules replaced one implementation per variant. See [ffl-scoring-rules.md](ffl-scoring-rules.md)
- [x] FFL historical import **tooling** — human-in-the-loop capture (userscript → ingest → parse) feeding a layered pipeline (fixtures → squads → trades → submitted teams), with scores recomputed via per-season `Rules` and every score found kept in notes as a reference. Replaces the one-time-CLI approach originally scoped. Running the import is Phase 26. See [ffl-historical-import.md](ffl-historical-import.md)
- [x] Deletion semantics fast-follow (ADR-020 step 1) — `ffl.player_match.club_match_id` from `ON DELETE CASCADE` to `RESTRICT`, so a fixture delete slipping past the `hasTeams` guard fails loudly instead of silently destroying submitted teams. See [ADR-020](../ai/decisions/adr-020-deletion-semantics.md)

Emerged during the phase rather than being scoped up front — byes and superbyes were deferred from slice 2 and added as-we-go, and the rest came out of using the fixture builder:

- [x] Match model unification — one match = a style (`versus` / `bye` / `superbye`) plus its participating `club_match` rows, replacing the home/away pair. Byes and superbyes render on round and match pages; a bye counts toward For only, not games played
- [x] Fixture builder — staged round-by-round editor, round-robin generator plus manual rounds, superbyes, saves reconciled in place rather than rebuilt; rounds ordered by AFL round rather than creation order
- [x] Add Season rework — explicit rules-era selection and club picker at creation time
- [x] Admin / Data Ops separation — data-ops split into its own package and schema, Admin (Seasons, Fixtures) split out of Data Ops, Calculate tab moved to Data Ops, nav decluttered

Not-found and player navigation, arising from use rather than plan:

- [x] Not-found handling — single-entity queries resolve unknown and unparseable ids to null instead of erroring (`pgx.ErrNoRows` → `domain.ErrNotFound` through repo and resolver), a shared `NotFound` page across the FFL and AFL views, and a catch-all route for unmatched paths. Remaining unreached call paths and the typecheck gate are Phase 28
- [x] Player navigation — header player search jumping to any player's season page, and a player's other seasons linked as year + club chips from their season page
