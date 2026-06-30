---
name: next-phase
description: Close the current phase and open the next one. Marks the current phase done in the roadmap, moves it to history, replaces current-sprint.md with the next phase tasks, and creates a new git branch. Triggered by /next-phase.
---

Close the current phase and open the next one.

## Steps

### 1. Identify current and next phase

Read `plans/roadmap.md` to find:
- The current phase (lowest-numbered phase without ✅)
- The next phase (the one after it)

If the current phase still has unchecked `[ ]` items, stop and report them to the user before proceeding. Do not close a phase with open tasks unless the user confirms.

### 2. Check the current branch

Run `git branch --show-current`.

- **If it starts with `phase-`**, extract the phase number from it:
  - If that number matches the phase you're about to open (N+1), you're already on the right branch — skip branch creation in step 7 and note in the report that work continues on the existing branch.
  - If it does not match, stop and ask the user: "You're on `<branch>` (Phase <N_current>) — has this been merged into main?"
    - If yes: run `git checkout main && git pull`, then ask whether to create a new branch for Phase N+1 off main. Only create it (step 7) if they confirm.
    - If no: ask how they'd like to proceed (e.g. finish/merge it first, branch off it anyway, or continue Phase N+1 work on it) before creating anything new.
- **If it doesn't start with `phase-`**, no special handling is needed — proceed normally; the new branch will be created in step 7.

### 3. Confirm with the user

Show the user:
- Current phase number and title
- Any unchecked items (if none, say so)
- Next phase number, title, and goal
- The branch you'll end up on (either the existing branch from step 2, or the new branch name — see step 7)

Ask: "Ready to close Phase N and open Phase N+1?"

Do not proceed until the user confirms.

### 4. Mark the current phase done in roadmap.md

In `plans/roadmap.md`, add ✅ to the current phase heading. Example:
- `## Phase 21: UX — Navigation` → `## Phase 21: UX — Navigation ✅`

### 5. Move the closed phase to history.md

Append the full phase block (heading + all items) to `plans/history.md`, after the last existing entry. Use a blank line between entries.

Then remove the phase block from `plans/roadmap.md`.

### 6. Replace current-sprint.md

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

### 7. Create a git branch

Skip this step if step 2 determined you're already on the correct `phase-N-...` branch.

Otherwise, create a new branch named `phase-<N>-<short-slug>`, where N is the new phase number and the slug is a kebab-case summary of the phase title. Example: `phase-21-ux-navigation`.

The new branch must be cut from `main` (not from a stale phase branch). Make sure you're on an up-to-date `main` first — step 2 already handles this when moving off a merged phase branch; otherwise run `git checkout main && git pull` before branching.

```bash
git checkout -b phase-<N>-<short-slug> main
```

Confirm the branch was created and report it to the user.

### 8. Report

Summarise what was done:
- Phase N closed and moved to history
- current-sprint.md updated for Phase N+1
- Branch: `phase-N-slug` (created fresh off main, or continuing on the existing branch — say which)