# musiccat — Agent Specification

## Overview

`musiccat` is a Go CLI application to catalogue a personal music collection.

### Key Goals

- Fast, low-friction data entry
- Easy correction and undo
- Offline-first; local SQLite DB is the source of truth
- Future-proof for optional Discogs integration
- Incremental, reviewable commits from agents

### Required Reading

Agents must read:

1. `docs/shared/AGENTS.md` → global shared agent instructions
2. `docs/shared/CODING_STANDARDS.md` → coding standards
3. `docs/shared/CONTRIBUTING.md` → commit and contribution guidelines
4. `AGENTS.md` → repo-level agent rules
5. `docs/agent-brief.md` (this file) → project-specific instructions

---

## Fixed Decisions

- **Language:** Go
- **CLI framework:** Cobra
- **Database:** SQLite at `~/.musiccat/musiccat.db`
- **Format state storage:** config file (`~/.musiccat/config.toml`)
- **Recent default:** 10 entries
- **App/User-Agent name:** `musiccat`

## Architectural Principles

### Data Storage

- Local SQLite database is canonical
- No cloud sync
- No background services

### MusicBrainz Integration

- MusicBrainz release groups are canonical
- `musicbrainz_release_group_id` is unique when not NULL
- Can be NULL for manual entries

### Minimal Metadata

- Do not store track-level data
- Minimal cached metadata:
  - artist name
  - release title
  - first release year (optional)
  - MusicBrainz release group ID

### Discogs (Future)

- Discogs optional
- `discogs_release_id` on ownership records only
- Not used in Phase 1

### Release Types

- Do not store release type (albums vs singles)
- MusicBrainz release group type may be displayed for disambiguation

### Genres

- No genre system
- No genre storage or queries

## Database Schema

### `releases`

| Field | Type | Notes |
|-------|------|-------|
| `id` | INTEGER PRIMARY KEY | |
| `artist` | TEXT NOT NULL | |
| `title` | TEXT NOT NULL | |
| `year` | INTEGER | nullable |
| `musicbrainz_release_group_id` | TEXT UNIQUE | nullable |
| `created_at` | TEXT | default CURRENT_TIMESTAMP |

### `ownership`

| Field | Type | Notes |
|-------|------|-------|
| `id` | INTEGER PRIMARY KEY | |
| `release_id` | INTEGER NOT NULL | FK → releases.id |
| `format_category` | TEXT NOT NULL | CD, Vinyl, Tape, Digital |
| `format_detail` | TEXT | nullable (7", 12", LP, etc.) |
| `purchase_date` | TEXT | nullable |
| `cost` | REAL | nullable |
| `source` | TEXT | nullable |
| `notes` | TEXT | nullable |
| `discogs_release_id` | INTEGER | nullable |

**Notes:**

- Each physical/digital copy is a separate ownership row
- No quantity field

## Configuration

- **Config directory:** `~/.musiccat/`
- **Config file:** `config.toml`

**Stores:**

- Current batch format (CD, Vinyl, Tape, Digital)

## CLI Commands

### 1. `musiccat set-format <FORMAT>`

- Sets current batch format
- Persists in config file
- Used implicitly by `add`

### 2. `musiccat add "<artist name>"`

- Search MusicBrainz for artists
- User selects correct artist
- Fetch release groups for that artist
- Display: title, year, release group type
- User selects one or more release groups
- Insert into `releases` if not present
- Insert into `ownership` using current batch format
- No confirmation prompts
- Mistakes handled via `recent` / `undo`

### 3. `musiccat add --manual`

For releases not in MusicBrainz:

- Prompt for:
  - artist
  - title
  - year (optional)
  - format category
  - format detail (optional)
- Set `musicbrainz_release_group_id` = NULL
- Insert into both tables

### 4. `musiccat list`

- List all stored releases
- Optional filters: artist, format
- Offline-only

### 5. `musiccat update "<artist>" "<title>"`

- Locate matching ownership rows
- Prompt user if multiple matches
- Update:
  - `purchase_date`, `cost`, `source`, `notes`
  - `format_category`, `format_detail`
  - release fields: `artist`, `title`, `year`, `musicbrainz_release_group_id`

### 6. `musiccat recent [--format <FORMAT>]`

- Show last 10 (default) ownership additions
- Optional format filter
- Display ownership IDs for undo

### 7. `musiccat undo <ID | all>`

- Delete ownership entries by ID
- `all` undoes the most recent batch
- Optional confirmation when multiple rows deleted

## Error Handling

- Do not silently swallow errors
- Display meaningful messages
- Prefer recoverable workflows

## MusicBrainz API

- No account or API key required
- User-Agent required:
  ```
  User-Agent: musiccat/0.1.0 (your-email@example.com)
  ```
- Respect 1 request/sec rate limit
- Only fetch release group metadata as needed

## Implementation Constraints

- **Language:** Go
- **CLI:** Cobra
- **SQLite:** `github.com/mattn/go-sqlite3`
- **HTTP:** `net/http` or `go-resty`
- **JSON:** `encoding/json` or `go-resty`
- **Code must compile at every commit**

## Commit Discipline

- One concern per commit
- Each commit must compile

**Suggested progression:**

1. Project skeleton + dependencies
2. Config handling (format state)
3. SQLite schema + DB bootstrap
4. MusicBrainz artist search
5. Release group listing
6. `add` command
7. `add --manual`
8. `recent` and `undo`
9. `list`
10. `update`

## Non-Goals

- Web UI / TUI
- Genre system
- Track-level cataloguing
- Automatic Discogs sync
- Price valuation / collection worth estimates