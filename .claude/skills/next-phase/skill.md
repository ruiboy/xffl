---
name: next-phase
description: Close the current phase and open the next one. Marks the current phase done in the roadmap, moves it to history, replaces current-sprint.md with the next phase tasks, opens a MR to main with the phase, and — once the user merges — cuts the next phase branch from a freshly pulled main. Triggered by /next-phase.
---

Close the current phase and open the next one.

The branch model is fixed: **every phase lives on its own `phase-<N>-…` branch, is merged to `main` via a MR, and the next phase is always cut fresh from an up-to-date `main`** — never branched from a stale phase branch. So closing a phase always ends with a MR the user merges, and opening the next one happens after that merge.

## Steps

### 1. Identify current and next phase

Read `plans/roadmap.md` to find:
- The current phase (lowest-numbered phase without ✅)
- The next phase (the one after it)

If the current phase still has unchecked `[ ]` items, stop and report them to the user before proceeding. Do not close a phase with open tasks unless the user confirms.

### 2. Confirm with the user

Show the user:
- Current phase number and title
- Any unchecked items (if none, say so)
- Next phase number, title, and goal

Ask: "Ready to close Phase N and open Phase N+1?"

Do not proceed until the user confirms.

### 3. Mark the current phase done in roadmap.md

In `plans/roadmap.md`, add ✅ to the current phase heading. Example:
- `## Phase 21: UX — Navigation` → `## Phase 21: UX — Navigation ✅`

### 4. Move the closed phase to history.md

Append the full phase block (heading with ✅ + all items) to `plans/history.md`, after the last existing entry. Use a blank line between entries.

Then remove the phase block from `plans/roadmap.md`.

### 5. Replace current-sprint.md

Delete the content of `plans/current-sprint.md` and replace it with a sprint doc for the next phase. Format:

```
# Current Sprint — Phase N: <Title>

**Sprint goal:** <goal from roadmap, verbatim>

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [ ] <task 1>
- [ ] <task 2>
...
```

If the next phase in the roadmap only has a goal statement (no task list), note that in the sprint doc and prompt the user: "Phase N has no task list yet — refer to `plans/ideas.md` and confirm scope before starting work."

### 6. Commit the close-out and open a MR to main

The current phase's work is on its `phase-<N>-…` branch. Commit the close-out edits onto that branch so the phase's history rides in with the MR:

```bash
git add plans/roadmap.md plans/history.md plans/current-sprint.md
git commit -m "Close Phase <N>: <title>"
git push
```

Then open (or update) a MR to `main` for the branch — invoke `/create-mr`, or use `gh pr create` — so the whole phase lands on `main` in one review. Report the MR URL.

(Edge case: if the phase work isn't on a `phase-` branch — e.g. you're already on `main` — skip the MR and just commit the close-out; the next-phase branch is still cut from `main` in step 7.)

### 7. Hand off for merge, then branch from main

**Stop and ask the user to merge the MR.** Do not create the next-phase branch until it is merged — the next phase must start from a `main` that already contains this phase.

Once the user confirms the merge:

```bash
git checkout main && git pull
git checkout -b phase-<N+1>-<short-slug> main
```

The slug is a kebab-case summary of the next phase title (e.g. `phase-25-ffl-scoring`). Confirm the branch was created and report it.

### 8. Report

Summarise what was done:
- Phase N closed and moved to history; MR opened (and, once merged, landed on main).
- current-sprint.md updated for Phase N+1.
- Branch: `phase-<N+1>-slug`, cut fresh from an up-to-date main.
