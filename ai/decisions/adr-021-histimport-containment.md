---
status: accepted
date: 2026-07-20
scope: architecture
enforceable: true
rules:
  - "all historical-import code lives under an `histimport` package in each layer it touches (`application/histimport`, `infrastructure/histimport`); importer logic does not leak into general-purpose packages"
  - "dependencies point one way: `histimport` may import domain, existing infrastructure (e.g. the `forum` parser) and application use cases (e.g. `ImportRoundTeams`); no production package may import `histimport`"
  - "the only allowed inbound reference is the interface layer (GraphQL resolvers) wiring importer mutations/queries; that wiring is kept isolated so the whole tree can be excised in one cut"
---

# ADR-021: Historical-Import Code is Contained in `histimport` Packages, Excisable in One Cut

## Context

Phase 25 builds the FFL historical-import tooling (squad importer, spreadsheet
fixture importer, submitted-teams importer + delta view) and Phase 26 runs it across
2006–2025. See `plans/ffl-historical-import.md`.

This code is **one-off by design**. Its job is to backfill twenty seasons of forum
and spreadsheet data into the operational tables, then stop mattering. Once Phase 26
closes, the parsers for a defunct Tapatalk forum, the pasted-spreadsheet reader, the
author→club_season registry and the reconciliation delta view have no ongoing product
role. The historical-import doc already states as much for the coverage tracking
("it does not earn a permanent product surface").

The risk is the ordinary one for throwaway tooling that lives in a shared codebase: it
quietly grows roots. A general repository starts calling an importer helper, a domain
type gains a field only the importer needs, a shared package imports a parser — and the
"one-off" is now load-bearing and cannot be removed without regression risk. The
existing `forum` and `spreadsheet` infrastructure packages, plus `application/dataops`,
already mix genuinely-shared capture code with import-only code, which is exactly how
that entanglement starts.

We want the opposite property: at the end of the backfill era, deleting the import code
should be a mechanical excision, not an archaeology project.

## Options

1. **Leave importers in the existing packages** (`forum`, `spreadsheet`, `dataops`),
   distinguished only by file naming. Rejected: nothing prevents production code from
   depending on them, so the excision boundary is unenforceable and erodes by default.

2. **A separate Go module / service for the importers.** Rejected: over-heavy. The
   importers need in-process access to the domain and to `ImportRoundTeams`; a module
   split buys isolation we can already get with package boundaries and costs a build,
   wiring and cross-boundary-call tax for tooling that is temporary anyway.

3. **Dedicated `histimport` packages per layer, with a one-way dependency rule.**
   Chosen.

## Decision

### All import code lives under `histimport`

Historical-import code goes in an `histimport` package in each layer it touches:

- `internal/application/histimport/` — orchestration for the three importers and the
  reconciliation/delta-view queries.
- `internal/infrastructure/histimport/` — the import-only parsers (squads-thread
  parser, spreadsheet season-sheet parser) and their `testdata/`.

Import-only parsing does **not** go in the general `forum`/`spreadsheet` packages. The
existing `forum` parser (the four team-submission formats) and the DataOps forum-capture
tab are genuinely shared and stay where they are; `histimport` depends inward on them.
The empty `infrastructure/spreadsheet` package is removed — the season-sheet reader is
import-only and belongs under `histimport`.

### The dependency arrow points one way

`histimport` may import the domain, existing infrastructure (the `forum` parser,
postgres repositories) and existing application use cases (`ImportRoundTeams`,
`buildFFLSeason`). **No production package may import `histimport`.** A domain type,
repository or shared use case that finds itself needing something from `histimport` is a
signal the dependency is inverted — the shared thing moves out of `histimport`, or the
importer does the adapting on its own side.

The single permitted inbound reference is the **interface layer**: the GraphQL resolvers
that expose the importer mutations and queries reference `application/histimport`. That
is the allowed Interface → Application direction (ADR-005), and it is kept isolated
enough — its own resolver files, its own schema section — that the resolvers come out
with the rest of the tree.

### The excision test

The operative test for the boundary: **deleting the `histimport` packages and their
interface wiring must leave the production build compiling and every non-import code
path unchanged.** If removal breaks a production package, the one-way rule has been
violated and the leak is fixed rather than tolerated.

## Consequences

- The import tooling can be removed in one mechanical cut once Phase 26 closes, with a
  compiler-checked guarantee that nothing operational depended on it.
- Some duplication is accepted deliberately: if the importer needs a variant of a shared
  helper, it copies rather than reaches for a production package's internals or forces a
  shared package to depend on it. Duplicated throwaway code is cheaper than a root.
- New importer formats and parsers in Phase 26 land inside `histimport` by default; the
  boundary is where they live, not a thing to re-decide per parser.
- The rule is greppable and enforceable in review: any `import` of a `histimport` path
  from outside `histimport` or the interface layer is a defect.
