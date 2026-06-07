# Current Sprint — Phase 22: Real Data Load & Seed Cleanup

**Sprint goal:** Replace synthetic seed data with real 2026 AFL and FFL data, and slim the dev seed to a lean scaffold.

See `plans/ideas.md` for full item descriptions where relevant.

---

## Tasks

- [ ] Import all 2026 AFL rounds played to date (teams, players, match stats) via existing import tooling
- [ ] Import all 2026 FFL round teams to date
- [ ] Verify ladder calculation produces correct standings
- [ ] Smoke-test team submission and substitution flows against real data; capture edge cases
- [ ] Reduce `dev/seed` to a single representative round (enough for demo and future dev)
- [ ] Confirm e2e tests pass against slim seed
