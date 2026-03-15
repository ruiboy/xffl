---
name: checkdoc
description: Check if documentation needs updating based on recent changes
disable-model-invocation: true
---

Check if documentation should be updated based on code changes.

## Steps

1. Ask the user what scope to check:
   - **Uncommitted changes** — `git diff HEAD`
   - **Current branch** — `git diff main...HEAD` (all commits since branching from main)
   - **Entire codebase** — review all code against all documentation
2. Gather the changes for the chosen scope and summarise what they affect.
3. Ask the user: "These changes affect [brief summary]. Should I check if docs need updating?"
4. If the user says yes:
   - Review `README.md`, files in `doc/`, and files in `ai/` (excluding `ai/prompts/`)
   - For each file, check if the changes in scope make it outdated or incomplete
   - Suggest specific, minimal edits — keep documentation lean
   - Show the proposed changes and wait for approval before applying
5. If the user says no, done.