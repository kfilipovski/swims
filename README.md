# swims

A CLI for querying the [USA Swimming Data Hub](https://data.usaswimming.org) and persisting swim data in a local [Dolt](https://www.dolthub.com/repositories) database — a MySQL-compatible, version-controlled SQL database.

Search for swimmers, sync their times, query results with flexible filters, and visualize progression with ASCII graphs.

## Prerequisites

- [Go](https://go.dev/dl/) 1.22+
- [Dolt](https://docs.dolthub.com/introduction/installation) installed and on PATH

## Install

```bash
go build -o swims .
```

## Usage

### Initialize the database

```bash
./swims init
```

### Search and track swimmers

```bash
./swims sync search --first "Jane" --last "Doe"
./swims swimmers list
```

### Sync times

```bash
./swims sync times --person-key 896236    # sync a specific swimmer
./swims sync times --all                  # sync all tracked swimmers
```

### Query times

```bash
# By swimmer name (case-insensitive substring match)
./swims times Doe --course SCY

# Best time per event
./swims times Doe --best --course LCM

# Filter by event, age, meet, date range
./swims times Doe --event "200 BR" --age 14
./swims times Doe --meet "Sectionals" --since 2024-01-01

# Filter by competition season (Sep 1 – Aug 31) or calendar year
./swims times Doe --season 2025
./swims times Doe --year 2024

# Sort by time or power points instead of date
./swims times Doe --sort time
./swims times Doe --sort points

# Compare multiple swimmers
./swims times Doe Smith --best --course SCY
```

### Visualize progression

```bash
./swims times Doe --event "200 BR" --graph time      # time progression
./swims times Doe --event "100 FR" --graph points     # power points progression
```

### Database summary

```bash
./swims status
```

### Global flags

| Flag | Description |
|------|-------------|
| `--data-dir` | Path to Dolt database directory (default: current directory) |

## How it works

Data is fetched from the USA Swimming Data Hub's Sisense JAQL API and stored in four normalized Dolt tables: `swimmers`, `events`, `meets`, and `times`. All writes go through `sync` commands; query commands are read-only.

Since Dolt is a version-controlled database, every sync creates a commit — you get a full history of how your tracked swimmers' times change over time, with the ability to diff and branch just like git.

## License

MIT
