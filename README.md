# musiccat

A Go CLI application to catalogue a personal music collection.

## Quick Start

### Installation

1. **Build the binary:**
   ```bash
   git clone <repo-url>
   cd musiccat
   go build -o mc .
   ```

2. **Install to your PATH** (optional):
   ```bash
   sudo cp mc /usr/local/bin/
   # Or for current user only:
   mkdir -p ~/.local/bin
   cp mc ~/.local/bin/
   # Add to PATH if needed: export PATH="$HOME/.local/bin:$PATH"
   ```

3. **Set up shell completion** (recommended):
   ```bash
   # For bash:
   mc completion bash > ~/.local/share/bash-completion/completions/mc
   source ~/.bashrc
   
   # For zsh:
   mc completion zsh > "${fpath[1]}/_mc"
   
   # For fish:
   mc completion fish > ~/.config/fish/completions/mc.fish
   ```

### First Steps

1. **Set your preferred format:**
   ```bash
   mc set-format CD
   ```

2. **Add your first release:**
   ```bash
   mc add "Blur"
   # Or use the short alias:
   mc a "Radiohead"
   # Use 00 for next page, 99 for previous page in artist search
   ```
   
   Select releases from the MusicBrainz results, mark as promo/pirate with `p`/`i` suffix.

3. **Add a manual entry (not in MusicBrainz):**
   ```bash
   mc add --manual
   # Press Enter to repeat last artist for batch entry
   # Answer 'y' to add another release
   ```

4. **List your collection:**
   ```bash
   mc list
   mc ls --artist Blur
   mc l --format Vinyl --tag bootleg
   ```

5. **Find missing releases:**
   ```bash
   mc missing "Radiohead" --album
   ```

6. **View statistics:**
   ```bash
   mc stats
   mc st
   ```

7. **Update an entry:**
   ```bash
   mc update 5
   mc u 10 --acquired-date 2024-06-15 --cost 12.99
   mc u --release-id 42 --year 1995
   ```

### Common Workflows

**Add multiple releases in batch:**
```bash
mc set-format CD Album
mc add "Artist Name"
# Select multiple: 1,2,3 or with modifiers: 1p,2i,3(2)
```

**Search and filter:**
```bash
mc list --artist "Stone Roses" --format Vinyl
mc list --promo --sort year
mc list --tag bootleg --desc
mc list --year 1995
```

**Add another copy of existing release:**
```bash
mc add --release-id 42
# See release ID in 'mc list' output (RelID column)
```

**Update release metadata:**
```bash
mc update --release-id 42 --year 1995
mc update --release-id 42 --artist "Various Artists"
```

**Fix format category mistakes:**
```bash
mc update 123 --format-category CD
# For bulk: for id in 123 124 125; do mc update $id --format-category CD; done
```

**Tag management:**
```bash
mc update 10 --tag bootleg,live
mc tag rename old-tag new-tag
mc tag delete unwanted-tag
mc port --pattern "bootleg|promo"  # Migrate notes to tags
```

**Sync MusicBrainz data:**
```bash
mc sync 50-65  # Update range of IDs
```

**Quick recent additions:**
```bash
mc recent
mc r
```

## Command Reference

### Commands
- `add (a)` - Search and add releases by artist
  - `--release-id <ID>` - Add another copy of existing release
  - `--manual` - Manually enter release details (with artist cache)
  - `--album`, `--single`, `--ep`, etc. - Filter by type
- `list (ls, l)` - List releases with filtering (shows ownership ID and release ID)
  - `--tag <name>` - Filter by tag
  - `--artist`, `--title`, `--year`, `--format` - Filter options
  - `--sort <field>` - Sort by id, artist, title, year, format, added
- `missing (m)` - Find releases you don't own for an artist
- `update (u, up)` - Update ownership or release details
  - `update [id]` - Update ownership entry
  - `--release-id <ID>` - Update release metadata (artist, title, year)
  - `--format-category <format>` - Change format category
  - `--tag`, `--remove-tag`, `--set-tag` - Manage tags
- `tag` - Tag management
  - `rename <old> <new>` - Rename tag
  - `delete <name>` - Delete tag
- `port` - Migrate notes patterns to tags
- `stats (st)` - Display collection statistics
- `recent (r)` - Show recently added items (with tags)
- `sync` - Sync MusicBrainz metadata
- `undo (un)` - Remove ownership entries (with confirmation for older entries)
- `set-format` - Set default format for batch adding

### Supported Formats
- **CD**: Album, Single, EP, Maxi (no "CD" prefix)
- **Vinyl**: 7", 10", 12", Album, LP
- **Cassette**: Album, Single, Tape
- **Digital**: Album, Single, EP

### Tags
- Tags are automatically canonicalized (lowercase, hyphenated)
- Add tags: `mc update 10 --tag bootleg,live`
- Filter by tag: `mc list --tag bootleg`
- Manage tags: `mc tag rename old new`, `mc tag delete unwanted`
- Migrate from notes: `mc port --pattern "bootleg|promo"`

### Item Modifiers
When adding releases, use these suffixes:
- `p` - Mark as promo (e.g., `14p`)
- `i` - Mark as pirate (e.g., `14i`)
- `(n)` - Add quantity (e.g., `14(2)` for 2 copies)
- Combined: `14ip(2)` - Pirate promo, 2 copies

### Data Storage
- Database: `~/.musiccat/musiccat.db` (SQLite with WAL mode)
- Config: `~/.musiccat/config.toml`
- Tables: releases, ownership, tags, ownership_tags
- Automatic schema migrations on first run
- MusicBrainz API for metadata

### Artist Search Navigation
- Enter number to select artist
- `00` - Next page (25 artists per page)
- `99` - Previous page

## Development

```bash
# Run tests
go test ./...

# Run specific test suite
go test ./cmd/helpers/...

# Build
go build -o mc .
```
