# FFL Historical Import (2006–2025)

Backfill historical FFL data from the Tapatalk forum (`ffltkf`), which holds
seasons **2006–2025** (one sub-forum per season), working **backwards from 2025**
and stopping wherever the source data runs out.

**Tooling is Phase 25; the import itself is Phase 26.** Phase 26 is not pure data
entry — each season may need importer changes as new formats appear, so it stays a
code phase with a data-entry track.

1998–2005 predate the forum and would need separate sourcing — out of scope.

## Core principle

**The forum says what teams did; scores are recomputed.** We have AFL stats for
2006–2025 and per-season `Rules`, so a player's fantasy score =
`Rules.Score(position, aflStats)`. Every score we find — forum-posted or from the
spreadsheet — is a **reference, not a source**: it goes to `club_match.notes` /
`player_match.notes` as `posted:NN`, never to `drv_score`, which stays derived from
`player_match` rows. (Writing a reference score into `drv_score` would be silently
wiped by `recalculateFFLClubMatchScore` anyway.)

## Sources

| Source | Gives |
|---|---|
| **Manual builders** (built) | season, club_seasons, rounds, fixtures, finals |
| **Spreadsheet** (pasted) | rounds + fixtures + club-level reference scores |
| **Forum** | squads and every submitted team, H&A and finals alike |

Squads and trades were once scoped as an XML import from external software; **no
such export exists**. Squads come from a forum thread instead — a separate parser
from the four team-submission formats in `forum/parser.go`. Trades are **always
manual**: low volume, inconsistent format, and `addFFLPlayerToSeason` /
`removeFFLPlayerFromSeason` already take `fromRoundId` / `toRoundId`, so there is no
write path to build.

## Per-season workflow

Repeated for each season, newest first:

1. **Season + clubs** — manual. Creates the season with its `rules_id` and a
   `ClubSeason` per club.
2. **Squads** for all clubs — parse the season's thread, else enter manually. Does
   *not* require rounds to exist; only trades do.
3. **Fixtures** for the minor rounds — spreadsheet paste where available, else the
   manual builder.
4. **One minor round of teams** for all clubs — including bye and superbye
   club_matches, which field teams like any other.
5. **Trades** — manual, via the squad page. Applied between rounds, so the incoming
   player's `PlayerSeason` window opens on the round they first play. A squad without
   its trades makes every team after the first trade window unresolvable.
6. **Repeat 4–5 for every minor round.**
7. **Verify after the last minor round** — ladder plus per-round, per-club scores.
   Report discrepancies in both.
8. **Finals** — create the rounds manually, import the teams.
9. **Verify finals** scores and results.
10. **Close the season**, move to the next.

### Why squads before teams
Resolving messy names ("LDU NM", "B King GCS") against ~800 AFL players is the
scariest problem. A squad collapses it to a **closed set** — a submitted team can
only name players on that club's squad that round (~25–30). Resolve each squad
member once per season; weekly teams then largely auto-resolve. It does not remove
fuzzy matching, it moves it from ~2,300 submitted teams to ~300 squad entries per
season.

## Phase 25 — remaining slices

All import code lives in `histimport` packages (`application/histimport`,
`infrastructure/histimport`) with a one-way dependency rule so it can be excised after
Phase 26 — see [ADR-021](../ai/decisions/adr-021-histimport-containment.md).

Slices 0–2 are done: scoring eras confirmed in `rules_eras.go`; capture (userscript
→ ingest → in-session preview); season creation and the fixture builder, including
byes, superbye and finals.

3. **Squad importer** — squads-thread parser + **bulk resolution-review UI**. The
   per-post `FflPlayerLinkModal` path is not it. Also needs a raw-text passthrough in
   the capture preview: `ForumCaptureBuffer.Ingest` currently discards raw HTML and
   keeps only parsed output, so an unparseable thread shows nothing to work from.
4. **Spreadsheet fixture importer** — pasted season sheet → rounds, fixtures, and
   reference club scores to `notes`. Fixtures are upstream of teams in the per-season
   workflow (step 3 before step 4), so the tool that builds them lands before the
   submitted-teams importer. 2025's clubs and full fixture are already entered by hand,
   so this is first *needed* for 2024, but it is built here.
5. **Submitted-teams importer** — commit via the existing `ImportRoundTeams`, plus:
   - a **season-scoped author→club_season registry**. This does *not* fall out of
     season setup: `buildFFLSeason` takes `clubIds` and never sees a forum author
     name, and `forum.TeamForAuthor` is a hardcoded four-author map with no season
     dimension. Without it, `parseFFLTeamSubmission`'s `teamName` is typed by hand
     once per post, ~2,300 times.
   - **post classification** — team submission vs banter; submitted vs scored.
   - **authoritative-post selection** — multiple posts per author per round (team,
     then scored, then edits); default to the latest scored post, allow override.
   - the **per-match evaluated-vs-reference delta view**. This ships here, not later:
     the eras in `rules_eras.go` are reviewed but never verified against data, and
     deltas are the only proof a season's `rules_id` is right. Importing a full season
     without them risks 22 rounds scored on the wrong era, undetected.

**No coverage dashboard is built.** Progress lives in the Phase 26 sprint doc,
cross-checked against committed data with SQL. Once Phase 26 closes it is not needed
again, so it does not earn a permanent product surface.

### Phase 26 progress table

One row per season, following the workflow steps above. The doc is the working
record; SQL against `match` / `player_season` / `player_match` is what confirms it.

| Season | Clubs | Squads | Fixtures | Rounds done | Ladder ✓ | Finals | Closed |
|---|---|---|---|---|---|---|---|
| 2025 | ✓ | | ✓ | 0 / 22 | | | |
| 2024 | | | | | | | |

Import across as many sessions as needed: done = in the database, so the committed
data is always the real answer if the table drifts.

## Reconciliation

Two independent references may exist per club_match — the forum-posted score and the
spreadsheet score, both in `notes`. `drv_score` stays the evaluated truth.

- **Per club_match** — evaluated vs each reference; surface deltas (parser bug?
  posted typo? AFL-stat mismatch?).
- **Ladder at the end of the minor round** — computed vs the spreadsheet's. This is
  the only end-to-end check of the **superbye extra-points** fold and bye handling,
  which are otherwise covered by unit tests alone. Treat it as a per-season gate.
- **Finals** — scores and progression.

Delta *policy* (what to do when evaluated ≠ posted) is decided case by case as they
appear, not up front.

## Reference

**Match model.** A match is a `match_style` (`versus` | `bye` | `superbye`) plus its
`club_match`es (each with a `side`): versus = 2, bye = 1, superbye = N. A bye scores
toward `For` only — not a played round, no premiership points. Superbye: each score →
`For`, top scorer(s) earn 1 extra point (`drv_extra_points`, ladder "EP" column);
ties share, all-zero awards none. Bye and superbye teams are entered through the same
Import / Mark Final flow as any club.

**Capture.** A userscript (`dev/userscripts/ffl-forum-capture.user.js`) reads the DOM
and POSTs to the local ingest endpoint via `GM_xmlhttpRequest` — a bookmarklet can't,
because `fetch()` from the https forum to localhost is blocked by mixed-content and
CORS. Navigation stays human: it only captures the page you are on, never walks
pagination. Captured posts are held **in-session only**; re-capture is one click, and
idempotency lives at the commit layer (`SetTeam` is diff-based, so re-importing a
round is always safe).

**Escape hatch.** An unknown author, format or player never crashes a run — it is
flagged for mapping or skip. Parsers are extended as new formats are met.

## Open questions
- 1998–2005 — separate sourcing, separate effort.
- How far back the forum threads and spreadsheets stay usable — discovered season by
  season, not decided up front.
