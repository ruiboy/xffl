# FFL Historical Import (2006–2025)

Phase 25, second task. Backfill historical FFL data from the Tapatalk forum
(`ffltkf`), which holds seasons **2006–2025** (one sub-forum per season). Scores
are **recomputed** from AFL data via each season's `Rules`; the forum is the
source of *what teams did*, not of scores.

1998–2005 predate the forum and will be sourced separately (email/paper/none) —
out of scope here.

## Core principles

- **The forum gives team selections and structure; scoring is recomputed.** We
  already have AFL stats for 2006–2025 and per-season `Rules`, so a player's
  fantasy score = `Rules.Score(position, aflStats)`. Posted forum scores are
  **kept, not trusted**: written to `club_match.notes` / `player_match.notes`
  (labelled `posted:NN`), used only for reconciliation. Evaluated vs posted
  deltas are a deliberate **second pass**.
- **Scaffolding before teams.** Fixtures, squads, and trades are imported before
  weekly submitted teams — this is a hard dependency (below), and it makes the
  hardest problem (player-name resolution) tractable.
- **Human-in-the-loop capture defeats bot detection.** No server-side scraping.
  A bookmarklet run in your authenticated browser captures each page; the backend
  never touches the forum.

## Prerequisite: nail down the scoring eras
Recomputed scores are only as correct as the `Rules`, and the current era
definitions in `rules_eras.go` are **approximations** — both the per-era
parameters (points, positions, bench, interchange) and the season each change took
effect. Before the import can produce trustworthy scores — and before
reconciliation against posted scores / the score-only spreadsheets means anything —
confirm:
- the **exact per-era parameters**, and
- the **exact season → era mapping**, then tag every `ffl.season.rules_id`.

This is a hard dependency: wrong eras → wrong recomputed scores → false
reconciliation deltas. Era model lives in [ffl-scoring-rules.md](ffl-scoring-rules.md).

## Forum artifact → schema (why the order is forced)

| Forum artifact | Becomes | Notes |
|---|---|---|
| **Fixtures** (who played who, which round) | `Match` + two `ClubMatch` | A submitted team is a `PlayerMatch` on a `ClubMatch` — no fixture, nothing to attach to. |
| **Squads** (a team's full player list) | `PlayerSeason` per `ClubSeason` | Also where clubs first appear per season (the team registry). |
| **Trades** (ins/outs + effective round) | `PlayerSeason.from_round_id` / `to_round_id` | A trade *is* closing one PlayerSeason and opening another at a round boundary — already modelled. |
| **Submitted teams** (weekly selection) | `PlayerMatch` (scored via `Rules`) | Needs both a `ClubMatch` and a resolvable `PlayerSeason`. |

**Import order: fixtures → squads → trades → submitted teams.**

### The payoff — squads-first solves player resolution
The scariest problem is resolving messy names ("LDU NM", "B King GCS") against
~800 AFL players. Squads-first collapses it to a **closed set**: a submitted team
can only name players on that club's squad that round (~25–30 players). Resolve
each squad member **once per season** (with trade windows); weekly teams then
auto-resolve against the squad. This de-risks the whole backfill.

### Volume works in our favour
Scaffolding is low-volume — fixtures ≈ 1 post/season, squads ≈ one per club/season,
trades ≈ periodic. Dozens of posts per season enable the expensive layer (~2,300
weekly submitted teams across all seasons), which then largely auto-resolves.

## Shared infrastructure

### Capture — userscript (primary)
A **userscript** (Violentmonkey / Tampermonkey / Greasemonkey), not a bookmarklet.
The reason: getting the data *out* to `http://localhost` is the hard part — a
bookmarklet's `fetch()` from the `https` forum is blocked by mixed-content + CORS.
A userscript's privileged **`GM_xmlhttpRequest`** (with an `@connect localhost`
grant) bypasses both, so it can read the DOM **and POST straight to the local
ingest endpoint — no copy/paste at all.**

- The forum's JSON-LD blob is **truncated** — ignore it. Full post bodies live in
  the DOM: `div.post` → `div.content` (with `<br>`/table markup), and the author
  is reliably attached (`POST_AUTHOR` / `span[itemprop=name]`).
- `@match` the forum; the script auto-runs on each page, injects a **"Capture this
  round"** button (and inline confirmation, e.g. "✓ round 3, 4 teams saved"),
  walks `div.post` for `{topicId, postId, author, timestamp, html}`, and POSTs the
  page payload to the ingest endpoint.
- **Navigation stays human** — you click through threads/pages in your real
  logged-in session, one page at a time. The script only captures the page you're
  on; it never fetches pages itself (auto-walking pagination would look robotic and
  risk tripping bot detection).
- **Multi-page threads**: capture each page; posts are staged by `(topicId, postId)`,
  so a round's pages accumulate and dedupe naturally.
- You confirm **season + round** at capture time (season pre-filled from the
  breadcrumb; round titles are freeform — e.g. "SheepDog Trials" — so not reliably
  parseable).

**Fallback (zero-install):** a bookmarklet that writes the same JSON payload to the
**clipboard**, pasted into a DataOps "capture" box. Same ingest, one paste per page —
for when the extension isn't wanted.

### Staging (ephemeral)
Captured posts are held **in-session only** — no staging table. If a session ends,
re-capture the page (cheap: one click in the userscript). Idempotency lives at the
**commit** layer: re-importing a round is diff-based (via `SetTeam`), so re-running
is always safe. Durable progress is the committed data, not the staging — see
Progress tracking below.

### Team registry + escape hatch
Forum author → FFL club, **per season** (clubs come and go: Grand Pooh Bears,
Buckleys, …). Falls out of the fixtures/squads step. An unknown author posting a
team-shaped post is **flagged for mapping or skip**, never crashes the run.

## Import layers

**Only submitted teams come from the forum.** The scaffolding (season, fixtures,
squads) is built manually or from structured sources — no fuzzy fixture/season
parsing. This is deliberate: fixtures especially need a **permanent** builder
because finals are constructed as-we-go *every* season, forever — not a one-off
backfill tool.

1. **Season** (manual) — enter a **season name**, pick the **scoring era**
   explicitly (from `fflRulesEras`, pre-selected from the AFL season's year but
   overridable), the **AFL season**, and the **clubs** (checklist of existing clubs
   by id). Creates the season with the chosen `rules_id` and a `ClubSeason` per club.
2. **Fixtures** (manual **builder page** in Admin) — a staged, round-by-round editor
   saved atomically via `saveFFLFixtures` (reconcile: create / replace / delete;
   rounds with submitted teams are locked/immutable). Per round you add enough
   matches to cover the clubs; any leftover club is a **scoring bye** (a single-sided
   `match_style='bye'` match + one `ClubMatch`, so the club still fields a team and
   its score counts toward the season aggregate — no premiership points). Tools:
   **Repeat rounds X–Y** (optionally reversing home/away, auto-incrementing the AFL
   round) to fill the H&A season, and **add finals rounds individually as they happen**.
   The **superbye** (the one match variant beyond regular home-vs-away) is still deferred.
   This page is a lasting product feature, used live each finals series.
3. **Squads** → `PlayerSeason` rows, resolved to AFL players once per season.
   - **XML import** for recent years (consistent format from external software) —
     preferred where available: it front-loads the *closed-set* squad that makes
     team-import resolution easy.
   - **As-we-go** fallback for older years: players are added on first appearance
     during team import, topped up via the existing squad-maintenance page.
   - Trades are `PlayerSeason` `from/to_round_id` windows, set via the squad page.
4. **Submitted teams** (forum) → `PlayerMatch` via the existing `ImportRoundTeams`.
   - Must **classify posts**: team-submission vs banter/analysis (verbose threads
     contain many non-team posts) — heuristic on position-section structure, then
     human confirm. And submitted (no scores) vs scored (per-player/total present).
   - Multiple posts per author per round (team, then scored, edits): capture all,
     default-select the latest/scored post as authoritative, review overrides.

## Scoring & posted scores
Reuse `ImportRoundTeams` (already wired): it writes `posted:NN` per player and the
posted total to notes, and gets the **evaluated** score from `SetTeam` → `Rules`.
So each club_match ends with: submitted team composition, evaluated `drv_score`,
and the posted score preserved in `notes`.

## Reconciliation (deferred second pass)
- **Per club_match**: evaluated `drv_score` vs posted (from notes) → surface deltas;
  decide handling later (parser bug? posted typo? AFL-stat mismatch?).
- **Final, end-to-end**: reconcile evaluated season/round totals against the
  **score-only spreadsheets** (an independent source of truth) after all imports.

## Progress tracking (derived, not maintained)
Progress is durable because **the committed data is the record** — not a checklist,
and not the sprint doc (far too coarse for ~20 seasons × ~22 rounds × clubs). A
**coverage dashboard** in DataOps derives status by querying the real tables:
- **Fixtures** present for a season/round? (`match` rows) — defines the expected cells.
- **Squad** imported for a club/season? (`player_season` rows)
- **Submitted team** present for a `club_match`? (`player_match` rows), with an
  evaluated `drv_score` and a posted score in `notes`?

"Missing" = no rows; "partial" = fixture but no team, or a team with unresolved
players / no posted score. Ephemeral staging is fine: you never lose your place
because done = in the database. Import across as many sessions as needed; reopen
the dashboard to see what's left.

## Reuse (already exists)
- `forum.Parser.Parse(teamName, post)` — 4 formats, author supplies the team name.
- `dataops.ImportRoundTeams` — posted→notes, evaluated via `SetTeam`/`Rules`.
- DataOps player-link UI (`FflPlayerLinkModal`) — basis for resolution review.
- Live-scoring write path exists; **season/fixture write-side persistence does not
  yet** (reads only) and is built in slice 2.

## Delivery slices
0. **Nail down the eras** (done) — confirmed per-era parameters in `rules_eras.go`;
   season→era tagging folds into slice 2 (season creation auto-assigns `rules_id`).
1. **Capture + inspect** (done) — userscript → ingest → in-session parse/view.
2. **Season + fixtures** — the scaffolding, built manually (no forum parsing):
   - **2a** write-side persistence (season, club_season, round, match, club_match
     creates: sqlc + repos + domain).
   - **2b** season creation (year + clubs → season with auto `rules_id` + club_seasons).
   - **2c** fixture builder page (rounds + matches; round-robin copy-fill; byes,
     superbye, finals added as-we-go).
3. **Squads** — XML import (recent years, closed-set) + as-we-go/manual top-up
   (older years); resolution against the closed squad set.
4. **Trades** — `PlayerSeason` from/to-round windows (via the squad page).
5. **Submitted-teams importer** — post classification + authoritative-post selection,
   commit via `ImportRoundTeams`, evaluated scoring + posted-to-notes.
6. **Coverage dashboard** (derived from imported data) + **reconciliation**
   (per-match deltas, then spreadsheet).

Prove the whole chain on **one season** (ideally a recent, well-formatted one)
before scaling to all 20.

## Open questions / deferred
- **Scoring byes are modelled** (single-sided `match_style='bye'` match + one
  `ClubMatch`); the club fields a team and its score counts toward `For` only — not
  a played round, and no premiership points.
- **One match model.** A match is a `match_style` (`versus` | `bye` | `superbye`,
  a typed always-populated discriminator) plus its participating `club_match`es
  (each with a `side`). Versus = 2, bye = 1, superbye = N. The ladder folds one
  loaded type (`ScoredClubMatch`) by style; the builder API is a uniform
  `matches: [{ style, clubSeasonIds }]` per round. Superbye scoring: each score → `For`
  (no played round), top scorer(s) earn **1 extra point** (`drv_extra_points`, folded
  into total premiership points; ladder "EP" column). Ties share; all-zero awards none.
- **Bye/superbye team entry is wired**: `FFLMatch` exposes `matchStyle` + `clubMatches`
  (all sides), and the DataOps "FFL Teams" tab lists every `club_match` in the round
  (with a bye/superbye badge), so those teams are entered via the same Import / Mark
  Final flow as any club.
- Delta policy (what to do when evaluated ≠ posted) — decided in the second pass.
- 1998–2005 — separate sourcing, separate effort.
- Per-layer format cataloguing — expand parsers as new formats/teams are met
  (escape hatch keeps a run from stalling on an unknown format).
