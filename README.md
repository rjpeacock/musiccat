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

1. **Add your first release:**
   ```bash
   mc add "Blur"
   # Or use the short alias:
   mc a "Radiohead"
   ```
   
   Select releases from the MusicBrainz results, mark as promo/pirate with `p`/`i` suffix.

2. **List your collection:**
   ```bash
   mc list
   mc ls --artist Blur
   mc l --format Vinyl
   ```

3. **View statistics:**
   ```bash
   mc stats
   mc st
   ```

4. **Update an entry:**
   ```bash
   mc update 5
   mc u 10 --acquired-date 2024-06-15 --cost 12.99
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
mc list --source "Record Store" --desc
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
- `list (ls, l)` - List releases with filtering
- `update (u, up)` - Update ownership details
- `stats (st)` - Display collection statistics
- `recent (r)` - Show recently added items
- `sync` - Sync MusicBrainz metadata
- `undo (un)` - Remove ownership entries
- `set-format` - Set default format for batch adding

### Supported Formats
- **CD**: Album, Single, EP, Maxi-Single, Promo, Digipak
- **Vinyl**: LP, 12", 10", 7", Single, EP, Picture Disc
- **Cassette**: Album, Single, Tape
- **Digital**: Album, Single, EP

### Item Modifiers
When adding releases, use these suffixes:
- `p` - Mark as promo (e.g., `14p`)
- `i` - Mark as pirate (e.g., `14i`)
- `(n)` - Add quantity (e.g., `14(2)` for 2 copies)
- Combined: `14ip(2)` - Pirate promo, 2 copies

### Data Storage
- Database: `~/.musiccat/musiccat.db` (SQLite)
- Automatic schema migrations on first run
- MusicBrainz API for metadata

## Development

```bash
# Run tests
go test ./...

# Run specific test suite
go test ./cmd/helpers/...

# Build
go build -o mc .
```
