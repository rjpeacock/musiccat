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
| `format_detail` | TEXT | nullable (7", 12", Album, Single, EP, etc.) |
| `acquired_date` | TEXT | nullable |
| `cost` | REAL | nullable |
| `source` | TEXT | nullable |
| `notes` | TEXT | nullable |
| `discogs_release_id` | INTEGER | nullable |
| `is_promo` | BOOLEAN | default FALSE |
| `is_pirate` | BOOLEAN | default FALSE |

### `tags`

| Field | Type | Notes |
|-------|------|-------|
| `id` | INTEGER PRIMARY KEY | |
| `name` | TEXT NOT NULL UNIQUE | Canonical tag name (lowercase, hyphenated) |

### `ownership_tags`

| Field | Type | Notes |
|-------|------|-------|
| `ownership_id` | INTEGER NOT NULL | FK → ownership.id |
| `tag_id` | INTEGER NOT NULL | FK → tags.id |
| UNIQUE(ownership_id, tag_id) | | Prevents duplicate tags |

**Notes:**

- Each physical/digital copy is a separate ownership row
- No quantity field
- Tags are canonicalized: lowercase, trimmed, spaces → hyphens, collapsed separators

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

- Search MusicBrainz for artists with pagination (25 per page)
- Artist selection: enter number, `00` for next page, `99` for previous page
- User selects correct artist
- Fetch release groups for that artist
- Display: title, year, release type (with secondary types if present)
- User selects one or more release groups
- Support for quantities: `1(2)` for 2 copies of item 1
- Support for promo marking: `1p,2,3p` for promo variants
- Support for pirate marking: `1i,2,3i` for pirate copies
- Optional variant notes when quantity > 1
- Insert into `releases` if not present
- Insert multiple `ownership` rows for variants
- No confirmation prompts
- Mistakes handled via `recent` / `undo`

**Flags:**

- `--exact`: Exact artist name matching (filter results client-side)
- `--album`: Show only pure studio albums (excludes compilations, live, soundtracks)
- `--single`: Show only singles
- `--ep`: Show only EPs
- `--compilation`: Show only compilations
- `--live`: Show only live albums
- `--soundtrack`: Show only soundtracks
- `--year <YYYY>`: Filter by release year
- `--title <string>`: Filter by title (partial match)
- `--page-size <int>`: Results per page (default: 50)
- `--sort <fields>`: Sort by fields (type, year, title)
- `--desc`: Reverse sort order
- `--pirate`: Mark all selected releases as pirate copies
- `--release-id <ID>`: Add another copy of an existing release (prompts for format detail, notes, promo, pirate)

### 2b. `musiccat add --release-id <ID>`

- Add another copy of an existing release without searching
- Display current release details (artist, title, year)
- Prompt for format detail, notes, promo status, pirate status
- Uses current format category from config
- Useful for adding variants or duplicates quickly

### 3. `musiccat add --manual`

For releases not in MusicBrainz:

- Prompt for:
  - artist (first time: type name; subsequent: press Enter to repeat last artist)
  - title
  - year (optional)
  - format category (shows suggested format details)
  - format detail (optional, with suggestions)
- After adding each release, prompt "Add another? (y/N)"
- Artist cache persists within session for batch entry
- Set `musicbrainz_release_group_id` = NULL
- Insert into both tables

**Use case:** Batch-adding multiple releases from the same artist (e.g., Various Artists compilations)

### 4. `musiccat list`

- List all stored releases with ownership IDs and release IDs
- Display columns: ID, RelID, Artist, Title, Year, Format, Detail, Promo, Pirate, Acquired, Importance, Notes, Tags
- Release ID (RelID) shows which ownership entries share the same release
- Optional filters: artist, title, year, format, tag
- Sorting by id, artist, title, year, format, format_detail, added (default: year)
- Offline-only

**Filters:**

- `--artist <name>`: Filter by artist (partial match)
- `--title <name>`: Filter by title (partial match)
- `--year <YYYY>`: Filter by year
- `--format <name>`: Filter by format (CD, Vinyl, Cassette, Digital)
- `--tag <tagname>`: Filter by tag
- `--promo`: Show only promo items
- `--source <name>`: Filter by source
- `--notes <text>`: Filter by notes content
- `--sort <fields>`: Sort by field
- `--desc`: Reverse sort order

### 5. `musiccat update [id]`

- Update ownership entry by ID (or most recent if no ID provided)
- Interactive mode: prompts for all editable fields
- Flag mode: only update specified fields

**Update ownership fields:**

- `--acquired-date <date>`: Update acquired date
- `--cost <amount>`: Update cost
- `--source <source>`: Update source
- `--notes <text>`: Update notes
- `--format-category <format>`: Change format category (CD, Vinyl, Cassette, Digital)
- `--format-detail <detail>`: Update format detail
- `--promo`: Mark as promo
- `--pirate`: Mark as pirate

**Tag management:**

- `--tag <tagname>`: Add tag (repeatable)
- `--remove-tag <tagname>`: Remove tag (repeatable)
- `--set-tag <tagname>`: Replace all tags (repeatable)

### 5b. `musiccat update --release-id <ID>`

- Update release metadata (affects ALL ownership entries for that release)
- Interactive mode: prompts for artist, title, year
- Flag mode:
  - `--artist <name>`: Update artist name
  - `--title <title>`: Update title
  - `--year <YYYY>`: Update year (0 to clear)

**Use cases:**
- Fix incorrect year on manually entered releases
- Correct artist name typos
- Update title metadata

### 6. `musiccat recent [--format <FORMAT>]`

- Show last 10 (default) ownership additions with tags
- Optional format filter
- Display: ID, Artist, Title, Year, Format, Format Detail, Tags
- Display ownership IDs for undo

### 7. `musiccat undo <ID | all>`

- Delete ownership entries by ID
- `all` undoes the most recent batch (last 10)
- Confirmation required when:
  - Deleting multiple rows (`all`)
  - Deleting single entry NOT in last 5 added (safety guard)
- Recent entries (last 5) can be quickly undone without confirmation
- Shows details before deleting non-recent entries

### 8. `musiccat tag <subcommand>`

**Subcommands:**

- `rename <old> <new>`: Rename a tag (updates all ownership records)
- `delete <tagname>`: Delete a tag (removes from all ownership records)

### 9. `musiccat port [--pattern <regex>] [--keep]`

- Migrate notes patterns to tags
- Default: migrates all non-empty notes to tags
- `--pattern <regex>`: Extract specific patterns from notes
- `--keep`: Keep extracted text in notes (default: remove)
- Creates tags from matches (canonicalized)
- Idempotent: can be re-run without duplicating tags

### 10. `musiccat missing "<artist name>"`

- Find releases you don't own yet for an artist
- Search MusicBrainz for artist (with pagination: 00/99)
- Fetch complete discography
- Compare against owned releases (by MusicBrainz release group ID)
- Display missing releases sorted by type → year → title
- Shows count: "Found 42/87 releases you don't own yet"

**Flags (same as `add`):**

- `--exact`: Exact artist name match
- `--album`: Only albums (excludes compilations, live, soundtracks)
- `--single`: Only singles
- `--ep`: Only EPs
- `--compilation`: Only compilations
- `--live`: Only live albums
- `--soundtrack`: Only soundtracks
- `--year <YYYY>`: Filter by year
- `--title <string>`: Filter by title

---

## Format Conventions

### Format Categories and Details

**CD**: Album, Single, EP, Maxi (no "CD" prefix)

**Vinyl**: 
- Singles: 7"
- EPs: 7" or 12"
- Albums: 12" or Album
- Maxi-Singles: 12"

**Cassette**: Album, Single, Tape, Cassette

### Format Detail Inference

- CD Singles → "Single"
- CD EPs → "EP"  
- CD Albums → "Album"
- CD Maxi-Singles → "Maxi"
- Vinyl Singles → "7""
- Vinyl EPs → "7"" or "12"" based on context
- Vinyl Albums → "12"" or "Album"

### Multi-Variant Support

- Multiple ownership entries can exist for the same release
- Use `format_detail` to distinguish variants (e.g., different pressings, colors)
- Use `notes` field for additional variant information
- Example: Same album on LP and Picture Disc variants

---

## Phase 2 — Usage, Inspection, and Insight

Phase 2 focuses on **using and inspecting the catalogue**, not expanding external integrations.

### In Scope

- Smarter listing and filtering with tags
- Stable ID-based workflows
- Sorting (default: year)
- Collection statistics
- Tag management (rename, delete, port from notes)

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
- Display tags at the end of each line

Example:

```
ID  Artist     Title        Year  Format  Detail  Tags
14  Aerosmith  Nine Lives   1997  CD      Album   grunge,promo
```

---

#### 1.2 Filtering

Optional flags:

- `--artist <string>` (partial, case-insensitive)
- `--title <string>` (partial, case-insensitive)
- `--year <YYYY>`
- `--format <FORMAT>`
- `--tag <tagname>` (canonical form)

Filters are composable (AND logic).

---

#### 1.3 Sorting

- Default sort: `year` (ascending)
- `--sort <fields>`: Sort by comma-separated fields (id, artist, title, year, format, format_detail, added)
- `--desc`: Reverse sort order

---

### 2. `musiccat update` (Enhanced)

Editable ownership fields:

- `acquired_date`
- `cost`
- `source`
- `notes`
- `is_promo`
- `is_pirate`
- `format_category`
- `format_detail`

Editable release fields (with confirmation):

- `artist`
- `title`
- `year`
- `musicbrainz_release_group_id`

Tag management:

- `--tag <tagname>`: Add tag (repeatable)
- `--remove-tag <tagname>`: Remove tag (repeatable)
- `--set-tag <tagname>`: Replace all tags (repeatable)

Updates are flag-driven or prompted.

---

### 3. `musiccat stats`

Outputs:

- Total items owned
- Count by format
- Promo vs non-promo count
- Total spend (sum of `cost`)

No external API calls.

### 4. `musiccat add` (pagination and sorting)

**Goal:** List release groups in a **predictable, user-friendly order** with graceful handling of large result sets.

**Default Sort Order:**

1. **Pure releases first:** Albums without secondary types (no compilations/live/soundtracks) appear before albums with secondary types
2. **Secondary:** First release year (ascending)
3. **Tertiary:** Title (alphabetical)

**Display Format:**

- Title, year, release type
- Secondary types shown in brackets: `[Album + Compilation]`, `[Album + Live]`

**CLI Flags:**

- `--exact`: Exact artist name matching (case-insensitive)
- `--album`: Show only pure studio albums (excludes compilations, live, soundtracks)
- `--single`: Show only singles
- `--ep`: Show only EPs
- `--compilation`: Show only compilations
- `--live`: Show only live albums
- `--soundtrack`: Show only soundtracks
- `--year <YYYY>`: Filter by release year
- `--title <string>`: Filter by title (partial match)
- `--page-size <int>`: Results per page (default: 50)
- `--sort <fields>`: Sort by fields (type, year, title)
- `--desc`: Reverse sort order
- `--pirate`: Mark all selected releases as pirate copies

**Pagination:**

- Default page size: 50 items
- User may fetch next page by entering `99`
- CLI must indicate page number and total items if known
**Pagination:**

- Display results in pages (default: 50 per page)
- Navigate with page numbers
- Enter selections or next/previous page

**Example CLI Flow:**

```
mc add "Louis Armstrong" --page-size 40 --sort type,year,title
Displaying 1–40 of 172 releases

Album: What a Wonderful World (1967)
Album: Hello, Dolly! (1964)
...
Single: Hello, Dolly! (1964)

Enter number(s) to select, or 99 to see next page:
```

- User may select multiple releases (e.g., `1,2,3p` for promo)
- Selected releases are inserted into `releases` and `ownership` tables

### 5. Tag System

**Tag Canonicalization:**

- Tags are stored in canonical form: lowercase, trimmed, spaces → hyphens
- Multiple separators collapsed: `foo--bar` → `foo-bar`
- Prevents duplicates: "Promo", "promo", and "PROMO" all become `promo`

**Tag Management:**

- `musiccat tag rename <old> <new>`: Rename a tag across all ownership records
- `musiccat tag delete <tagname>`: Delete a tag from all ownership records
- `musiccat port [--pattern <regex>] [--keep]`: Migrate notes patterns to tags
- `musiccat update --tag <tagname>`: Add tags to ownership records
- `musiccat update --remove-tag <tagname>`: Remove tags from ownership records
- `musiccat update --set-tag <tagname>`: Replace all tags with specified tags

**Tag Display:**

- Tags are displayed at the end of each line in `list` and `recent` commands
- Format: comma-separated, canonical form
- Example: `promo,grunge,bootleg`

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
- **SQLite:** `modernc.org/sqlite` (pure-Go, no cgo dependencies)
- **Database Mode:** WAL mode with 5-second busy timeout
- **Connection Pooling:** MaxOpenConns=10, MaxIdleConns=5
- **HTTP:** `net/http` or `go-resty`
- **JSON:** `encoding/json` or `go-resty`
- **Code must compile at every commit**

## Database Connection Management

- **WAL Mode:** Enabled for concurrent reads/writes
- **Busy Timeout:** 5 seconds to prevent lock errors
- **Connection Pooling:** Prevents deadlocks from nested queries
- **Query Pattern:** Collect-then-process to avoid N+1 queries with open result sets
- **Bulk Fetching:** Use IN clauses and maps for efficient tag/metadata loading

## Commit Discipline

- One concern per commit
- Each commit must compile
- Tests must pass before commit

## Non-Goals

- Web UI / TUI
- Genre system
- Track-level cataloguing
- Automatic Discogs sync
- Price valuation / collection worth estimates
