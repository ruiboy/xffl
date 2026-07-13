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

### Capture (bookmarklet)
- The forum's JSON-LD blob is **truncated** — ignore it. Full post bodies live in
  the DOM: `div.post` → `div.content` (with `<br>`/table markup), and the author
  is reliably attached (`POST_AUTHOR` / `span[itemprop=name]`).
- Bookmarklet walks `div.post`, extracts `{topicId, postId, author, timestamp, html}`
  for every post on the page, and writes a JSON payload to the **clipboard**
  (avoids the https-forum → http-localhost mixed-content/CORS problem entirely).
- You paste the payload into a DataOps "capture" box. One paste per page.
- **Multi-page threads**: paste each page; posts are staged by `(topicId, postId)`,
  so a round's pages accumulate and dedupe naturally.
- You confirm **season + round** at capture time (bookmarklet pre-fills season from
  the breadcrumb; round titles are freeform — e.g. "SheepDog Trials" — so not
  reliably parseable).

### Staging (ephemeral)
Captured posts are held **in-session only** — no staging table. If a session ends,
re-paste the page (cheap: one bookmarklet click + paste). Idempotency lives at the
**commit** layer: re-importing a round is diff-based (via `SetTeam`), so re-running
is always safe. Durable progress is the committed data, not the staging — see
Progress tracking below.

### Team registry + escape hatch
Forum author → FFL club, **per season** (clubs come and go: Grand Pooh Bears,
Buckleys, …). Falls out of the fixtures/squads step. An unknown author posting a
team-shaped post is **flagged for mapping or skip**, never crashes the run.

## Import layers (each its own parser + review step)

Each layer shares capture + staging but has its own format handling and review UI.
Formats vary within a layer too (the 4 current teams already post 4 different ways).

1. **Fixtures** → seasons (tagged `rules_id`), rounds, `Match`/`ClubMatch`, `ClubSeason`.
2. **Squads** → `PlayerSeason` rows; resolve each squad member to an AFL player once.
   - *Fallback when no clean squad post*: derive an approximate squad from the union
     of players a club submitted across the season; infer trade windows from
     first/last appearance. Not authoritative, but unblocks a season.
3. **Trades** → amend `PlayerSeason` `from/to_round_id` windows.
4. **Submitted teams** → `PlayerMatch` via the existing `ImportRoundTeams`.
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
- `application.ImportRoundTeams` — posted→notes, evaluated via `SetTeam`/`Rules`.
- DataOps player-link UI (`FflPlayerLinkModal`) — basis for resolution review.

## Delivery slices
0. **Nail down the eras** (prerequisite) — confirm exact per-era parameters and the
   season → era mapping in `rules_eras.go`; tag every `ffl.season.rules_id`.
1. **Capture + inspect** — bookmarklet, paste box, in-session parse + view of one
   pasted page. No commit, no staging table.
2. **Fixtures importer** — seasons/rounds/matches for one season; team registry.
3. **Squads importer** + closed-set player resolution (memoized per season) + review.
4. **Trades importer** — PlayerSeason windows.
5. **Submitted-teams importer** — post classification + authoritative-post selection,
   commit via `ImportRoundTeams`, evaluated scoring + posted-to-notes.
6. **Coverage dashboard** (derived from imported data) + **reconciliation**
   (per-match deltas, then spreadsheet).

Prove the whole chain on **one season** (ideally a recent, well-formatted one)
before scaling to all 20.

## Open questions / deferred
- Delta policy (what to do when evaluated ≠ posted) — decided in the second pass.
- 1998–2005 — separate sourcing, separate effort.
- Per-layer format cataloguing — expand parsers as new formats/teams are met
  (escape hatch keeps a run from stalling on an unknown format).
