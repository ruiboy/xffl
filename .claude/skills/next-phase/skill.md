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

### 2. Confirm with the user

Show the user:
- Current phase number and title
- Any unchecked items (if none, say so)
- Next phase number, title, and goal
- The branch name you'll create (see step 5)

Ask: "Ready to close Phase N and open Phase N+1?"

Do not proceed until the user confirms.

### 3. Mark the current phase done in roadmap.md

In `plans/roadmap.md`, add ✅ to the current phase heading. Example:
- `## Phase 21: UX — Navigation` → `## Phase 21: UX — Navigation ✅`

### 4. Move the closed phase to history.md

Append the full phase block (heading + all items) to `plans/history.md`, after the last existing entry. Use a blank line between entries.

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

### 6. Create a git branch

Create a new branch named `phase-<N>-<short-slug>` where the slug is a kebab-case summary of the phase title. Example: `phase-21-ux-navigation`.

```bash
git checkout -b phase-<N>-<short-slug>
```

Confirm the branch was created and report it to the user.

### 7. Report

Summarise what was done:
- Phase N closed and moved to history
- current-sprint.md updated for Phase N+1
- Branch created: `phase-N-slug`