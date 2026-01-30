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
- Support for quantities: `1(2)` for 2 copies of item 1
- Support for promo marking: `1p,2,3p` for promo variants
- Optional variant notes when quantity > 1
- Insert into `releases` if not present
- Insert multiple `ownership` rows for variants
- No confirmation prompts
- Mistakes handled via `recent` / `undo`

### 3. `musiccat add --manual`

For releases not in MusicBrainz:

- Prompt for:
  - artist
  - title
  - year (optional)
  - format category (shows suggested format details)
  - format detail (optional, with suggestions)
- Set `musicbrainz_release_group_id` = NULL
- Insert into both tables

### 4. `musiccat list`

- List all stored releases with ownership IDs
- Display format_category, format_detail, and notes
- Optional filters: artist, format, promo, source, notes
- Sorting by artist, title, year, format, added (default)
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

---

## Format Conventions

### Format Categories and Details

**CD**: Album, Single, EP, Maxi-Single, Promo, Digipak, Jewel Case

**Vinyl**: LP, 12", 10", 7", Single, EP, Picture Disc, Colored Vinyl

**Cassette**: Album, Single, Tape, Cassette

### Multi-Variant Support

- Multiple ownership entries can exist for the same release
- Use `format_detail` to distinguish variants (e.g., different pressings, colors)
- Use `notes` field for additional variant information
- Example: Same album on LP and Picture Disc variants

---

## Phase 2 — Usage, Inspection, and Insight

Phase 2 focuses on **using and inspecting the catalogue**, not expanding external integrations.

### In Scope

- Smarter listing and filtering
- Stable ID-based workflows
- Sorting
- Collection statistics

### Out of Scope

- Discogs enrichment or syncing
- MusicBrainz reconciliation
- Track-level metadata
- Bulk import workflows

---

## Phase 2 CLI Enhancements

### 1. `musiccat list` (Enhanced)

#### 1.1 Ownership ID Display

- Always display ownership IDs
- IDs must be usable with `update` and `undo`

Example:

ID Artist Title Year Format Promo
14 Aerosmith Nine Lives 1997 CD no

---

#### 1.2 Filtering

Optional flags:

- `--artist <string>` (partial, case-insensitive)
- `--format <FORMAT>`
- `--promo`
- `--source <string>`
- `--notes <string>`

Filters must be composable.

---

#### 1.3 Sorting

Flags:

- `--sort <field>`
  - `artist`
  - `title`
  - `year`
  - `format`
  - `added` (default)
- `--desc`

---

### 2. `musiccat update` (Restricted)

Only editable fields:

- `purchase_date`
- `cost`
- `source`
- `notes`
- `is_promo`
- `format_detail`

Non-editable by default:

- artist
- title
- year
- MusicBrainz release group ID

Updates must be flag-driven.

---

### 3. `musiccat stats`

Outputs:

- Total items owned
- Count by format
- Promo vs non-promo count
- Total spend (sum of `cost`)

No external API calls.

### 4. `musiccat add` (pagination and sorting)

Outputs:

- release groups listed in a **predictable, user-friendly order** 
- large result sets handled gracefully

Default Sort Order

- **Primary:** release type (`Album` → `EP` → `Single` → `Other`)
- **Secondary:** first release year (ascending)
- **Tertiary:** title (alphabetical)

CLI Flags

- `--sort <field>`: override default sorting
  - `type`, `year`, `title`
  - Can combine: `--sort type,year,title`
- `--desc`: reverse sort order

Pagination

- Default page size: 50 items
- User may fetch next page by entering `99`
- CLI must indicate page number and total items if known

Optional Filters

- `--album` / `--single`
- `--year <YYYY>`
- `--title <string>`: partial match on release title

### Example CLI Flow

mc add "Louis Armstrong" --page-size 40 --sort type,year,title
Displaying 1–40 of 172 releases

Album: What a Wonderful World (1967)

Album: Hello, Dolly! (1964)
...

Single: Hello, Dolly! (1964)

Enter number(s) to select, or 99 to see next page:

- User may select multiple releases (e.g., `1,2,3p`)  
- Selected releases are inserted into `releases` and `ownership` tables following existing Phase 1 logic

---

## Format Conventions & Ownership Notes

### Principles

- `format_category` is **hard-ish**: CD, Vinyl, Tape, Digital.
- `format_detail` is **semi-formal, convention-based**, filterable, and human-readable.
- Physical or content variants are recorded via `ownership` rows.
- Notes store additional distinguishing information, e.g., region, catalog number, special edition.

---

### Recommended `format_detail` by Category

#### CD

- `Album`
- `Single`
- `EP`
- `Maxi`

#### Vinyl

- `LP`
- `12"`
- `10"`
- `7"`

#### Tape

- `Album`
- `Single`

---

### Multi-Variant Ownership

- Multiple ownership rows may exist for the same release.
- Each row represents a **distinct physical or digital copy**.
- Variants are distinguished via:
  - `format_detail`
  - `notes`
- Example:

| release_id | format_category | format_detail | notes          |
|------------|----------------|---------------|----------------|
| 123        | CD             | Single        | UK CD1         |
| 123        | CD             | Single        | UK CD2         |
| 123        | CD             | Single        | JP bonus track |

- This allows tracking of multi-pack singles or promotional variants without adding new tables.

---

### CLI Guidance

- When adding releases, allow **multiple ownership rows** for the same release.
- Suggested `format_detail` values may be displayed, but **free text is always allowed**.
- Filtering, listing, and stats rely on consistent use of `format_category` + `format_detail`.

## Database Testing

### Use Generic Database Interfaces

- Write tests using `database/sql` interfaces only
- Don't depend on `go-sqlite3`-specific APIs
- Keep all logic generic

### Pure-Go Driver for Tests

Use a pure-Go driver for tests if sandboxed / cgo not available:

- Suggest using `modernc.org/sqlite` in test files
- Example: in `_test.go` files, import `_ "modernc.org/sqlite"` instead of `go-sqlite3`
- This lets tests run anywhere, including agent environments

### Use In-Memory Databases

Use temporary in-memory databases:

```go
db, err := sql.Open("sqlite", ":memory:")
```

**Benefits:**

- No file writes
- No state persistence needed
- Works for both drivers

### Test Behavior, Not Implementation

Test the important behavior, not the C implementation:

- Verify `releases` and `ownership` tables are created
- Verify insert/read/update/delete works
- Check constraints (e.g., unique MusicBrainz IDs)

### Skip Tests in Sandbox Environments

Skip DB tests if compilation fails:

- Wrap tests in `t.Skip()` if the driver cannot compile

```go
func TestDB(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping DB test in sandbox")
    }
}
```

This prevents the agent from blocking due to environment limitations.

### Use Table-Driven Tests

Use table-driven tests for common error scenarios:

- Duplicate MusicBrainz IDs
- Invalid formats
- Missing required fields

### Commit Discipline

Commit the tests separately:

- One commit for DB schema and helpers
- One commit for DB tests
- Make them small, reviewable, buildable

---

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

## Non-Goals

- Web UI / TUI
- Genre system
- Track-level cataloguing
- Automatic Discogs sync
- Price valuation / collection worth estimates
