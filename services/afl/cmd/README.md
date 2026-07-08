# AFL service binaries

This module builds **three** binaries from one shared core. Each `cmd/<name>`
is a separate entry point that compiles to its own artifact — the Go linker only
includes what that `main` reaches, so the service binary contains no import/
scraper code and vice versa. They are **separate deployments built from shared
source**, not one program.

All three are *driving adapters* over the same transport-agnostic core in
`internal/domain` and `internal/application` (see `ai/architecture/principles.md`
and ADR-005/016). The web service drives it over GraphQL/RPC; the CLIs drive it
from a terminal. There is exactly one place that knows the AFL domain.

| Binary | Path | Role | Talks to | Deploy shape |
|--------|------|------|----------|--------------|
| `afl` (service) | `cmd/main.go` | Serves the app | Postgres, event bus, HTTP clients | Long-lived service |
| `afltables-export` | `cmd/afltables-export` | Scrape historical stats → CSV | afltables.com (HTTP) | One-off / manual |
| `afltables-import` | `cmd/afltables-import` | Ingest historical CSV → DB | Postgres, stdin | One-off job (interactive) |

## `cmd/main.go` — the AFL service

The long-running service that backs the web app. On one HTTP port (`PORT`,
default `8080`) it serves:

- **GraphQL** — playground at `/`, queries at `/query` (federated via the router).
- **Twirp RPC** — `PlayerLookup` for cross-service calls (ADR-018).

It also opens a Postgres `LISTEN/NOTIFY` subscription (`shared/events/pg`) to
react to events. This is the only binary deployed to the serving fleet.

```
just run-afl                                   # local
DATABASE_URL=... PORT=8080 go run ./cmd/main.go
```

## `cmd/afltables-export` — historical scraper

Fetches player match stats from afltables.com and writes one CSV per season to
`afl-historical/<season>.csv`. Pure fetch → parse → CSV; it touches **no
database**. Polite (User-Agent, inter-request delay), re-runnable (overwrites),
walks `-from` down to `-to`.

```
just afl-historical-export 2023 2023
# or: cd services/afl && go run ./cmd/afltables-export -from 2023 -to 2023 -out ../../afl-historical
```

Parser + CSV writer live in `internal/infrastructure/afltables`.

## `cmd/afltables-import` — historical ingest (interactive)

Reads `afl-historical/<season>.csv`, creates the season→player_match scaffold,
and resolves player identity against existing `afl.player` records. Idempotent
(natural-key upserts + the `dataops_player_source` xref). Ingests **newest→
oldest** so a real career chains backward from the seeded modern seasons.

Resolution: an exact name adjacent to a known career auto-links; a brand-new
name auto-creates; and ambiguity prompts on **stdin** — a **season gap** (same
name, non-consecutive seasons → likely a different player), duplicate exact
names, or a high-confidence fuzzy near-match. Every new player and near-miss is
appended to a review log (`afl-historical/import-review.log`).

```
just afl-historical-import 1998 2023
# or: cd services/afl && DATABASE_URL=... go run ./cmd/afltables-import -from 1998 -to 2023
```

Because it prompts on stdin it is run attended, not as an unattended job. The
use case (`application.ImportHistoricalSeason`) and repo
(`postgres.HistoricalRepository`) are reused by the integration tests, which
drive the same core with an automatic prompter.
