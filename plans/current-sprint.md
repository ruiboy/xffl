# Current Sprint — Phase 26: FFL Historical Import (2006–2025)

**Sprint goal:** Run the Phase 25 importers across the seasons, backwards from 2025, stopping wherever the source data runs out. Not pure data entry — each season may need importer changes as new formats appear.

Per season, repeated: season + clubs → squads → minor-round fixtures → teams round by round, applying trades between rounds → verify ladder and per-round scores → finals → verify → close. See `plans/ffl-historical-import.md`.

---

## Tasks

- [ ] Progress table in the sprint doc, one row per season, cross-checked against committed data with SQL — no dashboard is built, the need ends with this phase
- [ ] 2025 first, proving the whole chain end to end before scaling
- [ ] Importer fixes as new squad/team/spreadsheet formats appear
- [ ] Reconciliation: per-match evaluated-vs-reference deltas, ladder at end of minor round, finals results
