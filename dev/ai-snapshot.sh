#!/bin/bash
#
# ai-snapshot.sh
# Snapshots the AI control plane context (architecture decisions, sprint plans,
# prompts, and Claude skills) into a single file for sharing or LLM ingestion.
#

# Resolve repo root (parent of dev/ where this script lives)
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUTPUT="$REPO_ROOT/ai-snapshot.out"

# Clear the file if it exists
> "$OUTPUT"

{
  echo "=== README.md ==="
  cat "$REPO_ROOT/README.md" 2>/dev/null
  echo -e "\n"

  echo "=== Directory Structure (ai) ==="
  tree "$REPO_ROOT/ai" 2>/dev/null
  echo -e "\n"

  echo "=== Directory Structure (plans) ==="
  tree "$REPO_ROOT/plans" 2>/dev/null
  echo -e "\n"

  echo "=== Core AI Files (Architecture, Plans, Decisions, Prompts) ==="
  find "$REPO_ROOT/ai/architecture" "$REPO_ROOT/plans" "$REPO_ROOT/ai/decisions/decisions.md" "$REPO_ROOT/ai/prompts" -type f -not -name "current-task.md" -exec echo "--- File: {} ---" \; -exec cat {} \; -exec echo -e "\n" \; 2>/dev/null
  echo -e "\n"

  echo "=== Claude Skills Structure ==="
  tree "$REPO_ROOT/.claude" 2>/dev/null
  echo -e "\n"

  echo "=== Claude Skills Content ==="
  find "$REPO_ROOT/.claude/skills" -type f -exec echo "--- Skill: {} ---" \; -exec cat {} \; -exec echo -e "\n" \; 2>/dev/null

} >> "$OUTPUT"

echo "AI context snapshot complete: $OUTPUT"
