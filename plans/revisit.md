# Revisit

Things to reconsider later. Not roadmap items — just thoughts to dump and come back to.

---

### Naming: "Roster"

Revisit whether "roster" is the right term for the list of players on an FFL club season. It currently refers to the 30-player pool a manager picks their weekly 22 from, but the word may not land right. Consider alternatives (squad, list, pool, etc.) and whether the term should change across the codebase or just in the UI.

---

### Federating the graph

Currently the two GraphQL services (AFL and FFL) are completely isolated behind a dumb reverse proxy. The frontend has to issue separate queries to each service and join client-side (e.g. the Roster page dual-queries FFL for roster data and AFL for stats). Consider whether Apollo Federation or a similar approach would be worth the complexity — it would let the frontend write a single query that spans both services, with the gateway handling the stitching. Trade-offs: simpler frontend queries vs. added infrastructure and coupling at the gateway layer.
