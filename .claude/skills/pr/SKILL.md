---
name: pr
description: Create or update a GitHub PR for the current branch with a one-sentence overview and bulleted change summary readable in under a minute
disable-model-invocation: true
---

Create or update a GitHub PR for the current branch.

## Steps

1. Run `git log main...HEAD --oneline` to see all commits on the branch.
2. Run `git diff main...HEAD --stat` to see which files changed.
3. Check whether a PR already exists: `gh pr view --json number,title,body 2>/dev/null`.
4. Read enough of the diff to understand what changed — focus on intent, not line-by-line detail.

## Writing the PR body

**Title:** one short phrase (under 60 chars), no period. Describes the feature or fix, not the implementation.

**Body structure:**

```
<One sentence. What this branch does and why — the "elevator pitch". No bullet points here.>

## Changes

- **<Area or concern>** — brief description of what changed and why. One sentence max.
- **<Area or concern>** — ...
```

**Rules for the body:**
- The opening sentence must stand alone: someone reading only that line should understand the branch.
- Each bullet covers a coherent area of change (e.g. **Domain model**, **GraphQL schema**, **Parser**, **Frontend**, **Tests**). Do not use one bullet per file.
- Bold the area heading; follow it with an em dash and a plain-English description.
- Omit file paths unless the path itself is the clearest way to identify something (e.g. a config file or migration).
- No "I added", "I changed" — just state the change.
- No implementation details that don't help a reviewer understand intent.
- The whole body should be readable in under a minute.

## Creating or updating

- **No existing PR:** run `gh pr create --title "..." --body "..."` using a HEREDOC.
- **PR exists:** run `gh pr edit --title "..." --body "..."` using a HEREDOC.

Return the PR URL when done.