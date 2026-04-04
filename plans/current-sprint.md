# Current Sprint

**Sprint goal:** Phase 7 — Data Model Refinements

Clean up AFL/FFL data models so relationships are correct, then propagate changes through domain entities, GraphQL schemas, frontend, seed data, and tests.

## Key decisions

- `afl.player.club_id` removed — players exist independently; attached to clubs via `player_season → club_season`
- `ffl.player.club_id` removed — same logic
- `afl.player_season` gains `from_round_id` / `to_round_id` (nullable; null = start/end of season)
- `afl.player_match` gains `status` (named/played/dnp)
- `ffl.player.name` → `drv_name` (every FFL player has an AFL player; name is derived)
- `ffl.player_match.score` → `drv_score`
- `ffl.player_match.status` stays underived (may be initialised from AFL, but takes its own values)
- `ffl.club_match` gets unique constraint `(club_season_id, match_id)`
- `ffl.match`: `home_club_match_id`/`away_club_match_id` nullable — versus: both set, bye: home only, else: `clubs` JSONB used
- `ffl.match.clubs` JSONB stores club_season_ids: `{"A": {"club_season_id": 2}, "B": {"club_season_id": 1}}`
- `AFLClub.players` GraphQL field dropped entirely
- No migration scripts — just update init SQL + seed data

## Tasks

### 1. AFL schema changes
- [x] Remove `club_id` from `afl.player`
- [x] Add `from_round_id` and `to_round_id` to `afl.player_season` (nullable FK → `afl.round`)
- [x] Add `status` to `afl.player_match` (varchar)
- [x] Drop index `idx_afl_player_club_id`
- [x] Add indexes for new columns

### 2. FFL schema changes
- [x] Remove `club_id` from `ffl.player`
- [x] Rename `ffl.player.name` → `drv_name`
- [x] Rename `ffl.player_match.score` → `drv_score`
- [x] Add unique constraint `(club_season_id, match_id)` on `ffl.club_match`
- [x] Drop index `idx_player_club_id`

### 3. Seed data
- [x] Update AFL seed to remove player `club_id` values, add `status` to player_match rows
- [x] Update FFL seed to match renamed columns

### 4. AFL domain + infrastructure
- [x] Remove `ClubID` from `Player` entity
- [x] Add `FromRoundID` / `ToRoundID` to `PlayerSeason` entity
- [x] Add `Status` to `PlayerMatch` entity
- [x] Update sqlc queries
- [x] Update repository implementations

### 5. FFL domain + infrastructure
- [x] Remove `ClubID` from `Player` entity (if present)
- [x] Keep `Name` on `Player` entity (DB column is `drv_name` but domain field stays `Name`)
- [x] Keep `Score` on `PlayerMatch` entity (DB column is `drv_score` but domain field stays `Score`)
- [x] Update sqlc queries
- [x] Update repository implementations

### 6. AFL GraphQL
- [x] Drop `AFLClub.play/mers` field and resolver
- [x] Add `status` to `AFLPlayerMatch` type
- [x] Update `updateAFLPlayerMatch` mutation input to include status
- [x] Regenerate + fix resolvers

### 7. FFL GraphQL
- [x] Update types if any field names changed (DB renames are transparent to GraphQL)
- [x] Regenerate + fix resolvers

### 8. Frontend
- [x] Remove any references to AFL player → club direct relationship
- [x] Update queries/components if field names changed
- [x] Verify all views still render correctly

### 9. Tests
- [x] Update AFL integration tests (remove club_id from player fixtures, add status)
- [x] Update FFL integration tests (renamed columns)
- [x] Run full test suite green
- [x] Run Playwright e2e tests green

## Out of scope
- FFL UX changes (nav bar, pages) — Phase 8
- Team composition rules — Phase 9
- Event integration (AFL→FFL) — Phase 11
