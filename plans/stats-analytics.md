# Stats & Analytics — Design (STATS-1)

**Status:** idea / design — not yet scheduled.
**Supersedes:** the old roadmap Phase 26 decision gate (revisit ADR-015: Typesense vs ClickHouse) and the former "CQRS player stats read model" idea (move stat reads to the search index) — both resolved by this design.
**Depends on:** Phase 25 (historical import) for the 20-year dataset; useful before it with the ~3 seasons already loaded.

**Question:** with 20 years of AFL + FFL history loaded, what is the best solution for records ("best ever FFL score", "most hitouts in a match"), longitudinal history ("total games, draws, average winning margin"), interesting quirks, and graphs?

**Short answer:** not a search index. Build a small **Stats read model in Postgres** (a `stats` schema of denormalised fact tables + precomputed record tables, rebuilt idempotently), exposed as a **Stats GraphQL subgraph** that federates onto the existing entities, charted in the frontend with **ECharts**, with **Metabase pointed at the same schema** for owner-side quirk-hunting. Keep Typesense strictly for text search. Defer ClickHouse indefinitely — the data will never be big enough to need it.

---

## 1. What "statistical digging" actually decomposes into

Four distinct query shapes, and they want different machinery:

| Shape | Examples | What it needs |
|---|---|---|
| **Records / leaderboards** | Best ever FFL score, most hitouts in a match, highest losing score | Top-N over a fact table, filterable by scope (all-time / season / club / player / era) |
| **Longitudinal aggregates** | Total games, draws, avg winning margin, win % by club by decade, head-to-head | GROUP BY + joins across match/club/season dimensions |
| **Quirks / exploration** | Longest streaks, biggest comebacks, "won every stat but lost", closest ladder finishes | Window functions (`LAG`, `RANK`, running sums), ad-hoc SQL, a human at a query console |
| **Graphs / time series** | Player form curves, club score history, ladder position over rounds, score distributions | Ordered series endpoints; percentile/histogram aggregates |

The third row is the tell: **streaks, deltas, and rankings are window-function territory**. No search engine does window functions. Whatever serves rows 1, 2 and 4 to the product UI, row 3 needs a SQL surface for the human doing the digging.

## 2. Scale reality check (this decides everything)

Dev DB at 2026-07 holds ~3 seasons: 26,496 `afl.player_match` rows, 639 matches. Extrapolating:

| Dataset | 20 years, est. rows |
|---|---|
| `afl.player_match` (the big one) | ~180–250k |
| `afl.match` / `club_match` | ~4.5k / ~9k |
| `ffl.player_match` | ~90–120k |
| **Total fact rows** | **< 500k** |

Half a million rows is not analytics scale — it is **smaller than the test datasets used to benchmark OLAP engines**. Postgres aggregates this in milliseconds with ordinary indexes; a full sequential scan of the entire fact set is single-digit milliseconds on any laptop. Every architectural decision below follows from this number: the problem is **modelling and API design**, not query horsepower.

## 3. Why the search index is the wrong primary tool

Typesense (and search engines generally) are document-retrieval engines: tokenise → match → rank. What they offer beyond that is facet counts and basic per-facet min/max/avg. They do not do:

- joins (player → match → opponent → era),
- window functions (streaks, running totals, rank-within-group),
- arbitrary GROUP BY with HAVING,
- percentiles/histograms for distributions,
- point-in-time correctness ("only final matches count").

You *can* index one document per player-match and get "top 10 by hitouts" via sort — but that's the only row of the table in §1 it covers, and you'd be maintaining a second copy of the data to get a worse version of `ORDER BY hitouts DESC LIMIT 10`. ADR-013's framing ("Typesense remains the read model for aggregated reads — season averages, rankings, top scorers") drew this boundary too generously, and notably the codebase has already quietly corrected it in practice: season averages and medians shipped as SQL (`AFLPlayerSeason.stats` backed by `PlayerSeasonStatsRepository`), not as Typesense queries. This design just makes that correction official.

**Typesense keeps its real job:** typeahead and full-text search ("find the player named Horne-something"), returning IDs into the graph. That part of ADR-015 stands.

## 4. Industry practice, mapped to this repo

The standard pattern at every scale is **separate the operational plane from the analytical plane**:

1. **Operational stores** stay normalised and serve transactions (the `afl.*`, `ffl.*` schemas).
2. An **analytical read model** is derived from them — denormalised, rebuildable, optimised for reads. At warehouse scale this is ELT into Snowflake/BigQuery/ClickHouse with dbt; at this scale it is Fowler's *Reporting Database* pattern: a schema of derived tables in the same Postgres.
3. A **semantic layer** names the metrics (what exactly is "winning margin"?) so every consumer — UI, BI tool, future LLM query planner — computes them the same way. At enterprise scale that's dbt metrics/Cube; here it's a well-documented view layer.
4. **BI tooling for exploration** (Metabase/Superset) sits on the read model for the humans; the product UI gets a purpose-built API.

This is also exactly CQRS applied at the data layer, which ADR-013 already endorses ("preserve the CQRS split for aggregated reads") — the only change is *what implements the read side*.

### Options assessed

| Option | Verdict | Why |
|---|---|---|
| **Postgres read model** (`stats` schema, derived tables/matviews) | ✅ **Recommended** | Full SQL power (windows, percentiles, joins); zero new infrastructure; fits the repo's existing "recalculate from scratch, idempotent" philosophy; trivially handles 100× the projected data |
| **Typesense as analytics store** | ❌ Reject for aggregates | §3. Keep for text search only |
| **ClickHouse** | ❌ Defer indefinitely | Columnar OLAP earns its ops cost from ~50–100M rows and high-cardinality scans. At 500k rows it is strictly slower to operate and no faster to query |
| **DuckDB** | 🟡 Optional side-tool | Superb for ad-hoc exploration (`ATTACH` the Postgres DB or query a parquet export directly, no server). Worth having in the toolbox for quirk-hunting sessions; not needed in the serving path |
| **Metabase (or Superset)** | ✅ Recommended for exploration | Point it read-only at the `stats` schema: instant ad-hoc charts, saved "quirk" questions, zero code. Most quirks are *found* interactively, then the good ones get promoted into the product. See Milestone 1 |
| **Cube.dev / dbt semantic layer** | ❌ Overkill | The metric registry can be a documented SQL view layer + GraphQL enum at this scale |

## 5. Recommended architecture

### 5.1 A Stats bounded context

Create a **Stats context** (a fourth service, or initially a module — see §8) that owns a `stats` schema. It is a *derived, disposable* read model: every table can be dropped and rebuilt from the operational schemas at any time. That property is what keeps it architecturally cheap — no migrations discipline, no backup obligation, no consistency ceremony beyond "rebuild".

**Cross-schema access needs an ADR.** ADR-003 forbids cross-schema access *between operational services*. The industry-standard carve-out is that the analytics plane reads the operational plane read-only (warehouses read replicas/CDC feeds, not service APIs). Recommend a new ADR permitting the Stats context **read-only SELECT on `afl.*` and `ffl.*`** (enforced with a dedicated PG role), on the grounds that: it writes only to `stats.*`; it can never influence operational behaviour; and the alternative — refetching 300k rows through GraphQL like the search reindex does — is ceremony without benefit. (If purity is preferred, the API-fetch route works at this scale; it's just slower and more code. The ADR should record the choice either way.)

### 5.2 The read model itself

Wide, denormalised fact tables — one row per grain, names resolved, no joins needed at query time ("one big table" is legitimate best practice at this scale):

```sql
CREATE SCHEMA stats;

-- Grain: one AFL player-match. ~250k rows.
CREATE TABLE stats.afl_player_match (
    afl_player_match_id INT PRIMARY KEY,
    player_id INT, player_name TEXT,
    club_id INT, club_name TEXT,
    opponent_club_id INT, opponent_name TEXT,
    season_id INT, season_name TEXT, round_id INT, round_name TEXT, round_seq INT,
    match_id INT, venue TEXT, start_dt TIMESTAMPTZ,
    kicks INT, handballs INT, disposals INT, marks INT, hitouts INT,
    tackles INT, goals INT, behinds INT, score INT,
    club_score INT, opponent_score INT, margin INT, result TEXT,  -- from the player's perspective
    match_final BOOL                                              -- correctness gate, see §6
);

-- Grain: one FFL player-match (fantasy score + position + activation status + era).
CREATE TABLE stats.ffl_player_match ( ... , scoring_era TEXT, ... );

-- Grain: one match (both codes). Margins, totals, results, streak inputs.
CREATE TABLE stats.match ( ... );

-- Precomputed records: the "Records" page reads this table directly.
CREATE TABLE stats.record (
    record_key TEXT,          -- e.g. 'ffl_highest_club_score', 'afl_most_hitouts_match'
    scope TEXT,               -- 'all_time' | 'season:<id>' | 'club:<id>'
    rank INT,                 -- keep top 10, not just #1 — ties and "previous record" for free
    value NUMERIC,
    holder_label TEXT, holder_entity TEXT, holder_id INT,
    match_id INT, season_name TEXT, achieved_on DATE,
    PRIMARY KEY (record_key, scope, rank)
);
```

On top: a **metric view layer** (`stats.v_club_season_summary`, `stats.v_head_to_head`, `stats.v_player_career`) that is the single place each metric is defined. `doc/useful.sql` is the embryo of this — promote its queries into versioned views instead of a scratch file.

### 5.3 Refresh strategy

Consistent with the repo's existing philosophy (ladders: "entire season recalculated from scratch — simpler and drift-free"):

- **Full rebuild** is the primitive: one idempotent job repopulates all of `stats.*` from the operational schemas. At 500k rows this is seconds. Expose as `just stats-rebuild` + a mutation.
- **Event-driven incremental**: subscribe (like the search service does) to `AFL.MatchUpdated(final)` and `FFL.MatchScoreFinalized` and rebuild the affected season slice. Because a lost event only means *stale stats until the next full rebuild*, the at-most-once delivery findings from the codebase review (`doc/review-findings.md` §6) are acceptable here — schedule a nightly full rebuild as the backstop and the staleness window is bounded. This is the correct consistency budget for analytics.
- Historical data (Phase 25 imports) needs no special path: it lands in `afl.*`/`ffl.*`, next rebuild picks it up.

### 5.4 API: a Stats GraphQL subgraph

Stats joins the federated graph — this is the architecturally elegant payoff of ADR-013:

```graphql
type Query {
  statsLeaderboard(metric: StatMetric!, grain: StatGrain!,   # e.g. FFL_SCORE × PLAYER_MATCH
                   scope: StatScopeInput, first: Int, after: String): LeaderboardConnection!
  statsRecords(category: RecordCategory, scope: StatScopeInput): [StatRecord!]!
  statsTimeseries(subject: TimeseriesSubjectInput!, metric: StatMetric!): [TimeseriesPoint!]!
  statsHeadToHead(clubA: ID!, clubB: ID!): HeadToHead!
}

# Federation: stats attach where users already are
type AFLPlayer @key(fields: "id") {
  id: ID! @external
  careerTotals: AFLCareerTotals!
  records: [StatRecord!]!
}
type FFLClub @key(fields: "id") { historySummary: ClubHistory! }  # total games, draws, avg margin…
```

Rules that carry over: metrics as **enums** (this *is* the semantic layer at the API boundary — one name per metric, defined once in the view layer); cursor Connections per ADR-014 for leaderboards; entity resolvers batched from day one (don't repeat the ADR-017 drift). This design also lands the ADR-015 "LLM query planner" vision cleanly: the planner populates `{metric, grain, scope}` against a closed vocabulary — far more tractable than generating SQL or search-engine queries.

### 5.5 Graphs

- **Product UI:** [Apache ECharts](https://echarts.apache.org/) via `vue-echarts` — the battle-tested Vue pairing; handles time series, distributions, heatmaps (the frontend already has a heatmap idiom to preserve), brush/zoom for 20-year series. Lighter alternative if only sparklines/bars materialise: Observable Plot or hand-rolled SVG. One ADR-011-style addition, added when the first chart ships.
- **Server sends series, client renders.** The `statsTimeseries` field returns ordered points; no client-side aggregation of raw facts. 20 years of per-round points for one player is ~500 points — no downsampling needed.
- **Exploration:** Metabase (Milestone 1 below). Promote discoveries into `stats.record` keys or curated views; the product's "Quirks" page can literally be a table of curated query results refreshed by the rebuild job.

## 6. Correctness concerns specific to this domain

These bite harder than any engine choice:

1. **Finality filtering.** Records must only be set by `data_status = final` matches (both axes for FFL). Bake it in structurally: the rebuild only ingests final matches, or every fact row carries `match_final` and every view filters on it. Never let a provisional 180-point score into "best ever".
2. **Era-dependent FFL scoring.** Phase 25 already plans a pluggable `ScoringStrategy` per season (e.g. star excluded hitouts from 2026). Two consequences: (a) historical FFL scores must be **stored as facts computed under their era's formula**, never recomputed with the current one; (b) every FFL fact row carries `scoring_era`, and cross-era leaderboards must either display the era or be segmented by it. "Most FFL hitout points ever" is a different question before and after a rule change — the model has to make that visible, not paper over it.
3. **AFL era context.** 20 years spans real-world rule/era changes (game length, interchange rules). Cheap insurance: a `stats.era` dimension keyed by season range, so views *can* segment even if v1 doesn't.
4. **Lineage.** Historical imports come from a different source (afltables CSV) than live rounds (FootyWire). Carry `source` on fact rows — when a record looks suspicious, the first question is always "which import produced this row".
5. **Identity over 20 years.** Player identity resolution across two decades (name changes, duplicates from fuzzy import matching) is the biggest data-quality risk in the whole plan. A `stats.v_data_quality` view (players with implausible gaps, duplicate name+DOB candidates, matches whose player scores don't sum to club score) should be part of the rebuild from day one — it will pay for itself during the Phase 25 backfill.

## 7. Data flow (end to end)

```
 afltables CSV (historical, once)   FootyWire (weekly)      Forum posts (weekly)
        │                                │                        │
        └── Phase 25 import CLIs ──► afl.* ◄─ Data Ops UI    ffl.* ◄─ Data Ops UI
                                        │                        │
                                        │   (read-only role, per new ADR)
                                        ▼                        ▼
                              ┌──────────────────────────────────────┐
                              │  Stats rebuild job                   │
                              │  full rebuild (nightly/manual)       │
                              │  + event-triggered season slices     │
                              └──────────────┬───────────────────────┘
                                             ▼
                                       stats.* schema
                                 (facts · views · records)
                                   │                  │
                     Stats GraphQL subgraph      Metabase (read-only)
                       (federated, :8083)          owner exploration
                                   │
                            Apollo Router
                                   │
                        Vue frontend + ECharts
                 (Records page · History pages · charts)
```

Typesense stays exactly where it is: text search in, entity IDs out — a sibling of this flow, not part of it.

## 8. Milestones

### Milestone 1 — Metabase exploration (cheap, do-now candidate)

Run Metabase OSS locally — free, self-hosted, fully offline — over a first cut of the `stats` schema (views over the ~3 seasons already loaded). Validates the metric definitions and answers quirk questions *before* any service is built. Product UI is out of scope.

**Decisions made:**
- **Edition:** Metabase Open Source (AGPL). No account, no license, no cloud. Anonymous telemetry disabled via `MB_ANON_TRACKING_ENABLED=false`.
- **Runtime:** Docker container alongside the dev stack in `dev/docker-compose.yml`. Port **3030** on the host (3000 clashes with Vite).
- **App data:** Metabase's own H2 file on a named volume (`metabasedata`). Single user — deleting the volume removes all trace.
- **DB access:** a dedicated **read-only role** (`stats_reader`), `SELECT` only, scoped to the `stats` schema. Metabase never sees `afl.*`/`ffl.*` directly.
- **Connectivity:** same compose network as `xffl-postgres`, connect via service name (or `host.docker.internal:5432`, the Apollo Router pattern).

⚠️ New infra dependency → needs an ADR line before it lands (rule 5 in CLAUDE.md) — fold into the stats-phase ADR, or a small standalone ADR if this milestone runs earlier.

**Tasks:**
- [ ] ADR coverage for Metabase as dev-stack infra
- [ ] Create `stats` schema and promote the first queries from `doc/useful.sql` into versioned views (e.g. `stats.v_club_season_summary`, `stats.v_head_to_head`) — SQL lives in `dev/postgres/init/`
- [ ] Create `stats_reader` role: `LOGIN`, `SELECT` on schema `stats` only; add to init SQL
- [ ] Add `metabase` service to `dev/docker-compose.yml` (image `metabase/metabase`, port 3030, `metabasedata` volume, telemetry off)
- [ ] First-run setup: local admin user; add Postgres connection using `stats_reader`
- [ ] Update docs: `just dev-up` output line + CLAUDE.md common-commands note (":3030 Metabase")
- [ ] Seed a few saved questions to prove the loop: highest FFL club score, biggest AFL winning margin, head-to-head grid
- [ ] Capture any metric-definition corrections back into the `stats` views (they become the semantic layer the subgraph reuses)

**Non-goals:** no embedding of Metabase charts in the Vue app (product charts come later via ECharts); no write access; no Metabase-side scheduled jobs.

### Milestone 2 — with Phase 25 (historical import)

Add the data-quality views (§6.5) and `source`/era columns as import lands. Import order: AFL history → FFL history (needs era scoring strategies first, as Phase 25 already sequences).

### Milestone 3 — the Stats subgraph (a future roadmap phase)

Resolve the old ADR-015 question as *"Typesense retained for text search only; analytics = Postgres stats read model; ClickHouse rejected at this scale"* — a revision to ADR-013's aggregated-reads boundary and ADR-015's scope, plus the new cross-schema-analytics ADR (§5.1). Build the Stats subgraph + rebuild job. Start it as a module inside an existing service only if a fourth service feels heavy — but the schema boundary (`stats.*`) and the read-only role are non-negotiable from day one, so extraction later is mechanical.

### Milestone 4 — product surfaces

Records page, club history pages, player career charts in the frontend; ECharts enters via a one-line ADR-011 amendment.

## 9. Triggers to revisit

Move to a columnar engine (DuckDB first, ClickHouse only with real ops appetite) only if one of these becomes true:

- fact grain drops below player-match (e.g. per-possession/event data → tens of millions of rows),
- interactive queries need sub-second scans over >20M rows,
- the rebuild job exceeds minutes and incremental complexity stops being worth it.

None are plausible for 20 years of AFL/FFL at player-match grain. The `stats.*` schema + GraphQL boundary means the engine swap, if ever needed, is a contained infrastructure change — the same argument ADR-015 already makes for search.
