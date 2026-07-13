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

### Raw post staging
Captured posts are stored raw and re-parseable, keyed by `(topicId, postId)`.
Re-running a parser never requires re-capture. Idempotent.

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

## Coverage tracking
A matrix per layer — season × round (× club) showing done / missing / failed /
partial — so the work survives many sittings and gaps are visible.

## Reuse (already exists)
- `forum.Parser.Parse(teamName, post)` — 4 formats, author supplies the team name.
- `application.ImportRoundTeams` — posted→notes, evaluated via `SetTeam`/`Rules`.
- DataOps player-link UI (`FflPlayerLinkModal`) — basis for resolution review.

## Delivery slices
1. **Capture + stage + inspect** — bookmarklet, paste box, raw staging, view parsed
   posts for one pasted page. No commit.
2. **Fixtures importer** — seasons/rounds/matches for one season; team registry.
3. **Squads importer** + closed-set player resolution (memoized per season) + review.
4. **Trades importer** — PlayerSeason windows.
5. **Submitted-teams importer** — post classification + authoritative-post selection,
   commit via `ImportRoundTeams`, evaluated scoring + posted-to-notes.
6. **Coverage matrix** + **reconciliation** (per-match deltas, then spreadsheet).

Prove the whole chain on **one season** (ideally a recent, well-formatted one)
before scaling to all 20.

## Open questions / deferred
- Delta policy (what to do when evaluated ≠ posted) — decided in the second pass.
- 1998–2005 — separate sourcing, separate effort.
- Per-layer format cataloguing — expand parsers as new formats/teams are met
  (escape hatch keeps a run from stalling on an unknown format).
