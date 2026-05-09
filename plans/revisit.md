# Revisit

Things to reconsider later. Not roadmap items — just thoughts to dump and come back to.

## Audit graph traversal shortcuts

Fields like `FFLPlayerMatch.player` let callers skip through `playerSeason → player`; `FFLPlayerMatch.playerSeasonId` is a bare FK when the object is already reachable. Audit all types in `query.graphqls` for relations that duplicate a traversal that already exists, then decide whether to remove the shortcut or keep it as a convenience alias with a deprecation note.

