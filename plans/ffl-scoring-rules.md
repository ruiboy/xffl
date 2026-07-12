# FFL Scoring — per-season rules

Phase 25. Make FFL scoring and team rules vary by season instead of hardcoded
constants, so historical seasons score with their own era's formula.

## Model

FFL "positions" are stat buckets (`goals, kicks, handballs, marks, tackles, hitouts,
star`), each scoring its stat(s) × a point value. Today every parameter is a package
constant in `domain/player_match.go` — which *is* the latest (2015+) era. Replace them
with a `SeasonRules` value object, one per era, selected per season.

- `SeasonRules` carries **everything**, including the parts that have never changed, so
  they're data rather than assumptions: per-stat point values, the positions + slot
  counts, each position's stat-set (the star's set is what gained/lost hitouts), bench
  size, interchange on/off, substitution rule.
- One generic scoring function replaces the `CalculateScore` switch — the star isn't
  special, it's just a position whose stat-set is longer. Eras become one-line deltas.
- Rules are **code** in the FFL domain (one value per era + a registry), version-
  controlled and unit-tested. A `ffl.season.scoring_rules_version` column stores which
  era a season uses; the **year→era mapping is data**, so exact years can be pinned later
  without code changes.

## One source of truth

`SeasonRules` drives scoring, team-selection validation, the team-builder UX, and a
human-readable "rules this season" description — all from the same data.

## Known eras (years approximate, confirm later)

- **1998** — goal 4, tackle 3, star includes hitouts, no interchange
- **1999** — hitouts removed from the star
- **2005** — tackle 4
- **2010** — goal 5
- **2015** — interchange position added

Each change applies to that season onward. A no-bench early era is suspected; bench size
is parameterized so it can be set when confirmed.

## Plan

1. **Refactor to parity** — extract today's constants into a single "current" `SeasonRules`
   + the generic engine; route all call sites through it (live scoring, bye scoring, team
   validation, and the forum parser's multipliers). Add the column, defaulted to current.
   Gate: 2026 scores identical before and after.
2. **Define the historical eras** as delta values + registry + a worked-example unit test
   per era.
3. **Backfill** (the FFL historical team import task) — tag each season with its era; score
   via its rules.

## Deferred

- Exact era years (data, set later).
- Whether bench size / sub rules actually vary across eras — parameterized either way.
- Substitution-rule modelling detail.
