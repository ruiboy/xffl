# Current Sprint — Phase 22: Real Data Load & Seed Cleanup

**Sprint goal:** Replace synthetic seed data with real 2026 AFL and FFL data, and slim the dev seed to a lean scaffold.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [x] Import all 2026 AFL rounds played to date (teams, players, match stats) via existing import tooling
- [x] Verify AFL ladder calculation produces correct standings
- [x] Calculate score(s) for bye players on bench; currently shows as "?".  Fix FFL Team import in Rounds 3/4.
- [x] Import all 2026 FFL round teams to date
- [x] Verify FFL ladder calculation produces correct standings
- [x] Smoke-test team submission and substitution flows against real data; capture edge cases
- [ ] Reduce `dev/seed` to a single representative round (enough for demo and future dev)
- [ ] Confirm e2e tests pass against slim seed
