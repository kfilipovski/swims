# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A Go CLI (`swims`) for querying the USA Swimming Data Hub (Sisense JAQL API) and persisting swim data in a local Dolt database (MySQL-compatible, version-controlled).

## Build & Run

```bash
go build -o swims .
./swims init                                          # initialize dolt DB + schema
./swims sync search --first "Jane" --last "Doe"       # search & save swimmers
./swims sync times --person-key 896236                # fetch & persist times
./swims sync times --all                              # sync all tracked swimmers
./swims swimmers list                                 # list tracked swimmers
./swims status                                        # show DB summary (counts)
./swims times Doe --course SCY                        # query times by swimmer name
./swims times Doe --best --course LCM                 # best time per event
./swims times Doe --event "200 BR" --graph time       # ASCII progression graph
./swims times Doe Smith --season 2025                 # compare swimmers, Sep 2024–Aug 2025
```

The `--data-dir` flag on any command overrides the Dolt DB location (default: cwd).

There are no tests in this project.

## Architecture

```
cmd/           Cobra CLI commands (root, init, sync, sync_search, sync_times, swimmers, times, status)
internal/
  usas/        USA Swimming API client (auth token, JAQL queries, response parsing)
  dolt/        Dolt CLI wrapper (shells out to `dolt` for SQL, add, commit, init)
  model/       Domain types: Swimmer, Time, Event, Meet, ParseEventCode()
  store/       DB operations: SwimmerStore, TimeStore (upsert, query, best)
  format/      tabwriter-based table output + ASCII graph rendering
```

**Data flow**: API → `usas.Client` → `model` structs → `store` upserts via `dolt.Dolt` wrapper → `dolt add` + `dolt commit`

**Database schema** (4 normalized tables in `internal/dolt/schema.go`):
- `swimmers` — tracked swimmers (swimmer_id PK)
- `events` — unique event codes with decomposed distance/stroke/course
- `meets` — unique meet names
- `times` — swim times referencing events and meets by foreign key; unique on `(swimmer_id, event_id, sort_key, swim_date, meet_id)`

**Key design decisions**:
- Dolt is accessed by shelling out to `dolt` CLI (`dolt sql -q "..." -r json`), not via go-mysql-driver
- The `.dolt/` directory lives in the project root (gitignored)
- All DB writes go through `sync` commands; `swimmers list`, `times`, and `status` are read-only
- SQL queries are built with `fmt.Sprintf` and `sqlStr()` (single-quote escaping), not parameterized
- Correlated subqueries don't work reliably in Dolt — best-time queries use JOIN with derived table instead
- `REPLACE INTO` used for idempotent upserts; events/meets use `INSERT IGNORE`
- Times are batched in groups of 50 for insert
- USA Swimming API auth uses a public token endpoint (no user credentials) — see `internal/usas/auth.go`

**Domain concepts**:
- `--season N` filters Sep 1 of year N-1 through Aug 31 of year N (competition year)
- `--year N` filters Jan 1 through Dec 31 (calendar year)
- `times` command accepts swimmer names as positional args (case-insensitive substring match on tracked swimmers)

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `dolt` — must be installed and on PATH