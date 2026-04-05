# Current Sprint

**Sprint goal:** Phase 9 — FFL Team Composition Rules

Define and enforce the rules for how an FFL team is structured each round.
Validation lives in the domain. The UI enforces the rules through good UX.

---

## Team Composition Rules (source of truth)

| Position  | Starters | Scoring |
|-----------|----------|---------|
| Goals     | 3        | 5 pts × goals |
| Kicks     | 4        | 1 pt × kicks |
| Handballs | 4        | 1 pt × handballs |
| Marks     | 2        | 2 pts × marks |
| Tackles   | 2        | 4 pts × tackles |
| Hitouts   | 2        | 1 pt × hitouts |
| Star      | 1        | all stats combined |
| **Total** | **18**   | |

**Bench** — up to 4 players:
- 1 backup star (no position pair; backs up the star slot)
- 3 dual-position players: each covers exactly 2 non-star positions; each non-star position may be covered by at most one bench player

**Interchange** — at most 1 position per match: the bench player covering that position may freely substitute the starter if they score higher.

**Substitution** — a bench player steps in only when the starter is DNP (did not play). A player who played but scored 0 cannot be substituted.

Teams need not be full.

---

## Backend Tasks

### B1 — Fix Score() for multi-starter positions ✅
**File:** `services/ffl/internal/domain/club_match.go`

`starters` is currently `map[Position]*PlayerMatch` — only one slot per position, last write wins. Change to `map[Position][]*PlayerMatch` so all starters score. Substitution and interchange logic applies per individual slot within the position group.

- [x] Change starters map type
- [x] Update substitution loop (iterate each slot per position)
- [x] Update interchange loop
- [x] Update total sum

### B2 — Add PositionSlots constants + ValidateTeam() ✅
**File:** `services/ffl/internal/domain/player_match.go`

- [x] Add `PositionSlots map[Position]int` (goals:3, kicks:4, handballs:4, marks:2, tackles:2, hitouts:2, star:1)
- [x] Add `ValidateTeam(entries []UpsertPlayerMatchParams) error` enforcing:
  1. Starter count per position ≤ PositionSlots[pos]
  2. Bench count ≤ 4
  3. At most 1 bench star (backupPositions contains "star")
  4. Non-star bench players have exactly 2 backup positions, none "star"
  5. Each non-star position in at most 1 bench player's backup pair
  6. At most 1 interchange position across all players
  7. Interchange position (if set) must be a recognised Position value

### B3 — Enforce validation in SetTeam command ✅
**File:** `services/ffl/internal/application/commands.go`

- [x] Call `domain.ValidateTeam(entries)` before opening the transaction; return error on failure

### B4 — Integration tests for all composition rules ✅
**File:** `services/ffl/internal/interface/graphql/integration_test.go`

- [x] Valid 18-starter team saves successfully
- [x] Too many starters for a position → error
- [x] 5 bench players → error
- [x] 2 bench stars → error
- [x] Bench player with 3 backup positions → error
- [x] Same position in two bench players' backup pairs → error
- [x] 2 interchange positions → error
- [x] Score() sums correctly across all starters in a multi-slot position

---

## Frontend Tasks

### F1 — Fix slot counts + state retention bug ✅
**File:** `frontend/web/src/features/ffl/views/TeamBuilderView.vue`

- [x] Fix positions array: marks→2, tackles→2, star→1
- [x] Add `initializedMatchId = ref<string|null>(null)`; guard the watch with `if (cm.id === initializedMatchId.value) return`
- [x] Update summary: "18 starters" not "22"

### F2 — Redesign bench UI ✅
**File:** `frontend/web/src/features/ffl/views/TeamBuilderView.vue`

- [x] Replace benchSlots ref with benchStarSlot + benchDualSlots
- [x] Render Backup Star row with interchange toggle
- [x] Render Bench 1/2/3 rows with dual position selectors and interchange toggle
- [x] Disable position pills already used across other bench dual slots
- [x] Allow at most 1 interchange active across all rows
- [x] Load existing team data into new structure in the watch initialiser
- [x] Update submitTeam() to serialise correctly

### F3 — Squad panel bench actions ✅
**File:** `frontend/web/src/features/ffl/views/TeamBuilderView.vue`

- [x] Replace B button with ★ and B actions
- [x] Disable ★ if benchStarSlot.player is set
- [x] Disable B if all 3 benchDualSlots are filled

### F4 — E2E tests for Team Builder ✅
**File:** `frontend/web/e2e/ffl-team-builder.spec.ts`

#### Layout and structure (read-only mode)
- [x] Club name shown as h1 heading
- [x] Manage button visible; Done button not visible
- [x] Position group headings present: Goals, Kicks, Handballs, Marks, Tackles, Hitouts, Star
- [x] Correct slot counts per position (3/4/4/2/2/2/1 = 18 total)
- [x] Bench section present with Backup Star row and 3 dual-position rows
- [x] No Remove buttons, no Squad panel, no position selectors in read-only mode

#### Layout and structure (manage mode)
- [x] Click Manage → Done button visible, Manage button gone
- [x] Squad panel visible in manage mode
- [x] Remove buttons visible on filled slots
- [x] Interchange toggles visible on bench rows
- [x] Dual-position selectors visible on bench dual-position rows
- [x] Click Done → returns to read-only (Manage visible, Done gone, Squad panel gone)

#### Empty team flow
- [x] Enter Manage → all slots still empty, full squad available in panel (tested via partial edit flow)
- [x] Done on empty team → returns to read-only with all slots still empty (no crash)

#### Partial edit → view → re-edit (state retention)
- [x] Enter Manage → add one player to Goals → click Done
- [x] In read-only mode: that player's name is visible in the Goals section
- [x] Click Manage again → player still in Goals slot (local state retained, not reset by Apollo)
- [x] Add a second player to Kicks → click Done
- [x] Both players visible in read-only mode

#### Navigate away and back (server state persisted)
- [x] Add a player to Goals → Done (saves to server)
- [x] Navigate away (e.g. click Squad in nav)
- [x] Navigate back to Team Builder
- [x] Player still visible in Goals slot in read-only mode (loaded from server)
- [x] Enter Manage → player still in slot

#### Continue building across multiple sessions
- [x] Load page with partially-saved team → enter Manage → existing players present
- [x] Add more players → Done → all players (old + new) visible in read-only mode
- [x] Enter Manage again → all players still present in correct slots

#### Bench: backup star
- [x] In manage mode: Squad panel shows ★ button per player
- [x] Click ★ on a player → appears in Backup Star row
- [x] ★ button disabled for other players once Backup Star is filled
- [x] Remove from Backup Star → ★ buttons re-enabled

#### Bench: dual-position
- [x] Squad panel shows B button per player
- [x] Click B → player appears in first empty dual-position bench row
- [x] Dual-position bench row shows two position selectors (empty initially)
- [x] Select position in first selector → that position disabled in other bench rows' selectors
- [x] B button disabled once all 3 dual-position slots are filled

#### Interchange
- [x] Interchange toggle visible on Backup Star row and each dual-position bench row
- [x] Click interchange toggle on one row → that row marked as interchange
- [x] Clicking interchange on a second row → first deactivates, second activates (only 1 active)
- [x] Interchange state preserved through Done → re-open Manage

---

## Execution Order

```
B2 → B1 → B3 → B4    backend: domain rules first, then enforce, then test
F1 → F2 → F3 → F4    frontend: fix basics first, then rebuild bench, then test
```

B2 can start immediately. F1 can start in parallel with B2 (slot counts are known).
F2 depends on F1. B4 depends on B1–B3.
