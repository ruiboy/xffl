---
status: accepted
date: 2026-07-20
scope: architecture
enforceable: true
rules:
  - "a parser or client that ingests an external source lives in a source-named infrastructure package (e.g. `infrastructure/afltables`, `infrastructure/footywire`, `infrastructure/forum`, `infrastructure/spreadsheet`); formats a source produces are files within that package, not sibling packages"
  - "the parser's port interface and its parsed DTOs live in the application layer beside the other ports (`application/ports.go`), and the infrastructure package implements them"
  - "import operations (resolve, match, write) are orchestrated in `application/dataops`, named by operation; an operation may draw on one or more source parsers"
---

# ADR-021: Importer Layout — Source-Named Infra Packages, Ports in Application, Orchestration in DataOps

## Context

The system ingests from several external sources: AFL stats (afltables, footywire),
FFL forum posts, a squads thread, forum discussion threads, and fixture spreadsheets.
More will follow. We want each new importer to be obvious in the code tree and to land
in a consistent place without re-deciding per parser.

An earlier version of this ADR contained all historical-import code in a single
`histimport` package tree with a one-way dependency rule, on the premise that the code
was throwaway and should be excisable "in one cut" after the backfill. That premise did
not survive contact with the importer list: AFL stats capture and the FFL team importer
are permanent, ongoing infrastructure, not throwaway. Sorting importers by *when they
might die* is the wrong axis, and it forced awkward constraints (deliberate duplication,
a ban on production code reusing a parser). It was discarded.

The AFL service already had the right pattern and it needed no ADR to justify it:
capture is organised by **source** — `infrastructure/afltables`, `infrastructure/footywire`.

## Decision

Importers are organised on two axes, one per layer:

- **Infrastructure = the source.** A parser or client that reads an external source lives
  in a package named for that source: `afltables`, `footywire`, `forum`, `spreadsheet`.
  Different *formats* a source produces are files inside that package — e.g. the four
  team-submission formats and the squads-thread parser are both `forum` files; the
  fixture-sheet reader is a `spreadsheet` file. Content type (`squads`, `fixtures`) is
  **not** an infrastructure package name — that would mix a content axis into a
  source-named layer.

- **Application = the operation.** The parser's port interface and its parsed DTOs sit in
  `application/ports.go` beside `TeamParser`/`ParsedPlayerRow`; the infrastructure package
  implements the port. Import *operations* — resolving parsed rows to records, matching,
  writing via use cases — are orchestrated in `application/dataops`, named by operation
  (`fflteams`, `squads`, `fixtures`), consistent with the operation-named files in AFL's
  application layer (`historical.go`, `live_round.go`). An operation may draw on more than
  one source parser.

This matches AFL's existing application layout (operation-named files) and its
infrastructure layout (source-named packages), so both services read the same way.

## Consequences

- A new importer is a predictable shape: a file in the relevant source package, a port in
  `application/ports.go`, and — when there is an operation to run — a file in
  `application/dataops`. Where it lives is settled, not re-litigated per parser.
- Nothing is firewalled by a lifecycle guess. If production code reuses a parser, that is
  allowed; it simply means the importer was not throwaway. Genuinely one-off tooling still
  stays self-contained (its parser file and its `dataops` operation), so removing it later
  is a localised deletion, not a global excision.
- The rule is greppable in review: a content-typed infrastructure package
  (`infrastructure/squads`), a parser port defined outside the application layer, or import
  orchestration outside `dataops` is a defect.
