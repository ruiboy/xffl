# Ideas

Aspirational candidates, designs, and things to revisit — not commitments. Items carry stable IDs
(`PAGE-*`, `STATS-*`, `REV-*`) where the roadmap or other docs reference them; larger designs live
in sibling files and are indexed here.

---

## UX

### Possible pages

| ID | Route | Status | What it shows | Effort |
|---|---|---|---|---|
| PAGE-2 | `/ffl/player-seasons/:fflPlayerSeasonId` | Expand-row in SquadView; not a full page | One club's ownership of a player: club-specific notes, averages scoped to "while they were mine", per-round breakdown. | S |
| PAGE-9 | `/ffl/afl/clubs/:aflClubId` | Unblocked (Phase 24 backfill) | All-time FFL history of an AFL club across seasons: players drafted, aggregate scoring, per-season breakdown. | — |
| PAGE-10 | `/ffl/afl/players/:aflPlayerId` | Unblocked (Phase 24 backfill) | All-time FFL career of an AFL player across seasons: FFL clubs, total * points, FFL games played. | — |

### Navigating the historical AFL data

The Phase 24 backfill loaded 1998–2023 AFL player/match stats, but nothing in the
webapp reaches it — AFL views only surface the live season(s). Make it navigable:

- **Season switching** on AFL views (round/match/ladder) so any season back to 1998 is reachable — a season picker or historical index. *(Shipped: season lives in the URL at `/afl/seasons/:id` + `/ffl/seasons/:id`, a season picker in the global nav, and the season ladder is the home page. See PAGE-11 below for the richer season landing.)*
- **AFL player career view across seasons** (delivers PAGE-10): full stat history, clubs, per-season averages/★ — now that players carry 20+ years of `player_season`/`player_match` rows.
- **AFL club season history** (delivers PAGE-9).
- Existing player/club links already exist on current pages; they just need a historical destination.

#### PAGE-11 — Season dashboard vs full ladder

Today the season landing *is* the full ladder. Split them: make the landing a glanceable
**season dashboard**, with the full ladder as its own page.

- **Dashboard** — a brief ladder (all teams; games / premiership-points / percentage
  columns only) plus a round snapshot: the current round if the season is live, else its
  last round — which for a completed season is the Grand Final, so a historic season
  headlines the premier.
- **Full ladder** — its own page with the complete table.

Symmetric for FFL.

### Player Notes

Fields in the schema, populated by backend processes, not yet rendered in the UI.

| Field | Written by | Proposed UX home |
|---|---|---|
| `FFLClubMatch.notes` | FFL import (score reconciliation rationale) | FFL match view — alongside club match score |
| `FFLPlayerMatch.notes` | FFL import (posted player scores) | FFL match view — alongside individual player scores |

---

## Stats & Analytics

| ID | Design | Summary |
|---|---|---|
| STATS-1 | [stats-analytics.md](stats-analytics.md) | Records, history, quirks, and graphs over 20 years of AFL+FFL data. Postgres `stats` read model (denormalised facts + precomputed records, idempotent rebuild) → Stats GraphQL subgraph federated onto existing entities → ECharts in the frontend; Metabase (local, free) over the same schema for owner exploration — Milestone 1 is a cheap do-now candidate. Typesense stays text-search-only; ClickHouse rejected at this scale. Supersedes the old Phase 26 decision gate and the former "CQRS player stats read model" idea. |

---

## Quality & Hardening

The cross-lens actions from the 2026-07 codebase review. Full detail, severity tags, and
`file:line` references: [`doc/review-findings.md`](../doc/review-findings.md) (dated snapshot at
commit `1ce978c`).

REV-1 (event reliability) graduated to roadmap Phase 26.

| ID | Action | Detail in |
|---|---|---|
| REV-2 | Give every derived field a rebuild path and a staleness check — re-derive `drv_result`, build score reconciliation as a drv-vs-recomputed diff; amend ADR-010 with both invariants. | review §6 |
| REV-3 | Emit `AFL.MatchUpdated(partial)` on final→partial reversal so FFL doesn't keep believing the match is final. | review §6 |
| REV-4 | Finish ADR-017: add the missing DataLoaders, convert direct resolver lookups, batch the federation entity resolvers. Collapses the worst N+1s. | review §5 |
| REV-5 | Reconcile the contradicting ADRs: retire ADR-008 properly; fix ADR-004's no-callback rule + stale facts; align ACL table naming across ADR-016/decisions.md/cookbook. | review §1 |
| REV-6 | Wire `MapPgError` (or delete it and amend ADR-009) — kills the raw-pgx-error leak. | review §3 |
| REV-7 | Fix test-tier violations: tag the Typesense container test; rewrite the pg dispatcher test hermetically; add search + shared to `just test-all`. | review §7 |
| REV-8 | Documentation truth pass: domain.md match-style fiction + `named` status; cookbook, testing.md, repo-map.md, frontend.md staleness. | review §2 |
| REV-9 | Frontend structural pass: settle the cross-feature import rule; split `TeamBuilderView.vue`; introduce fragments; scope Team Builder/Squad queries. | review §4 |
| REV-10 | Backup/restore parity guard: dump schema alongside data backups or diff live schema against init SQL. | review §3 |
| REV-11 | Replace name-based finals detection with a typed `round.round_type`; fixes the ladder finals leak (`CalculateLadder`/`FindFinalBySeasonID` filter on `data_status='final'`, not round type, so finals count toward the H&A ladder) and unblocks per-week finals nav. Latent until the Phase 24 backfill flipped 2024/2025 finals to `final`; the derivation SQL excludes them via `NOT ILIKE '%Final%'`, so the two now disagree. | [round-type-enum.md](round-type-enum.md) |

---

## Platform & Other

### Season setup imports

Once-per-season CLIs, run when a new season starts rather than every round.

- **AFL season player import** — once/season CLI; fuzzy name matching to existing players; accept/reject flow for new and retiring players.
- **FFL squad import** — once/season CLI; resolve FFL rosters to AFL player IDs.

### Player availability

Track AFL injury and suspension data per player season, and enforce availability at FFL team submission.

- Add availability columns to `afl.player_season`: `availability TEXT`, `available_from_round_id INTEGER`, `availability_note TEXT`
- Update init SQL and test-e2e init SQL; apply via `ALTER TABLE` to live DB per migration strategy in cookbook
- Import pipeline: pull AFL injury/suspension list; populate availability fields
- FFL team submission validation: reject naming an unavailable player
- UX: show availability status on free agents page (PAGE-4) and player season view (PAGE-1)

### Deployment

- CI-ready (GitHub Actions or similar)
- ADR — consider deployment options (AWS, GCP, etc.)

### Search UI

Full-text search view with filters (source, type), backed by Typesense as-is. (Index enrichment
for stats/aggregates and the ADR-015 Typesense-vs-ClickHouse question are covered by STATS-1.)

### Other

- **Live AFL data source** — afltables may serve for weekly reconciliation once historical load is complete, but a real-time feed would unlock live in-progress scoring
- **Mobile app**
- **Backup remote destination** — rclone supports S3, GCS, B2 with a single consistent interface; decide when deployment target is clearer
