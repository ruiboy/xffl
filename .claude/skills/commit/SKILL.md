---
name: commit
description: Commit changes and check if documentation needs updating
disable-model-invocation: true
---

Before committing, check if documentation should be updated based on the changes since the last commit.

## Steps

1. Run `git diff HEAD` to see all uncommitted changes.
2. Ask the user: "These changes affect [brief summary]. Should I check if docs need updating?"
3. If the user says yes:
   - Review `README.md`, files in `doc/`, and files in `ai/` (excluding `ai/prompts/`)
   - For each file, check if the current changes make it outdated or incomplete
   - Suggest specific, minimal edits — keep documentation lean
   - Show the proposed changes and wait for approval before applying
   - After applying (or if no doc changes needed), proceed to commit
4. If the user says no, proceed directly to commit.

## Commit rules

- Summarise the nature of the changes (new feature, bug fix, refactor, etc.)
- Keep the commit message concise (1-2 sentences) focused on the "why"
- Stage only relevant files — never stage `.env` or credentials