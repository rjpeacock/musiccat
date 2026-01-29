package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const dbFile = "musiccat.db"

func GetDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".musiccat", dbFile), nil
}

func OpenDB() (*sql.DB, error) {
	path, err := GetDBPath()
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Add busy timeout and WAL mode for better concurrency
	connStr := path + "?_busy_timeout=5000&_journal_mode=WAL"
	database, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, err
	}

	// SQLite best practice: limit to 1 connection to avoid SQLITE_BUSY
	database.SetMaxOpenConns(1)

	return database, nil
}

const createReleasesTable = `
CREATE TABLE IF NOT EXISTS releases (
    id INTEGER PRIMARY KEY,
    artist TEXT NOT NULL,
    title TEXT NOT NULL,
    year INTEGER,
    musicbrainz_release_group_id TEXT UNIQUE,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
`

const createOwnershipTable = `
CREATE TABLE IF NOT EXISTS ownership (
    id INTEGER PRIMARY KEY,
    release_id INTEGER NOT NULL REFERENCES releases(id),
    format_category TEXT NOT NULL,
    format_detail TEXT,
    acquired_date TEXT,
    cost REAL,
    source TEXT,
    notes TEXT,
    discogs_release_id INTEGER,
    is_promo BOOLEAN DEFAULT FALSE,
    is_pirate BOOLEAN DEFAULT FALSE
);
`

const createTagsTable = `
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);
`

const createOwnershipTagsTable = `
CREATE TABLE IF NOT EXISTS ownership_tags (
    ownership_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    UNIQUE(ownership_id, tag_id),
    FOREIGN KEY(ownership_id) REFERENCES ownership(id),
    FOREIGN KEY(tag_id) REFERENCES tags(id)
);
`

func BootstrapDB(db *sql.DB) error {
	if _, err := db.Exec(createReleasesTable); err != nil {
		return err
	}
	if _, err := db.Exec(createOwnershipTable); err != nil {
		return err
	}
	if _, err := db.Exec(createTagsTable); err != nil {
		return err
	}
	if _, err := db.Exec(createOwnershipTagsTable); err != nil {
		return err
	}
	// Migration: add is_promo column if missing
	_, err := db.Exec("ALTER TABLE ownership ADD COLUMN is_promo BOOLEAN DEFAULT FALSE")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	// Migration: add is_pirate column if missing
	_, err = db.Exec("ALTER TABLE ownership ADD COLUMN is_pirate BOOLEAN DEFAULT FALSE")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	// Migration: rename purchase_date to acquired_date
	_, err = db.Exec("ALTER TABLE ownership RENAME COLUMN purchase_date TO acquired_date")
	if err != nil && !strings.Contains(err.Error(), "no such column") && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}

// OwnershipInput holds all fields needed to insert a new ownership record.
type OwnershipInput struct {
	ReleaseID      int64
	FormatCategory string
	FormatDetail   *string
	AcquiredDate   *string
	Cost           *float64
	Source         *string
	Notes          *string
	DiscogsID      *int
	IsPromo        bool
	IsPirate       bool
}

// OwnershipUpdate holds all fields that can be updated on an ownership record.
type OwnershipUpdate struct {
	FormatDetail *string
	AcquiredDate *string
	Cost         *float64
	Source       *string
	Notes        *string
	DiscogsID    *int
	IsPromo      *bool
	IsPirate     *bool
}

func InsertRelease(db *sql.DB, artist, title string, year *int, mbid *string) (int64, error) {
	// First, check if release already exists
	if mbid != nil {
		var existingID int64
		err := db.QueryRow("SELECT id FROM releases WHERE musicbrainz_release_group_id = ?", mbid).Scan(&existingID)
		if err == nil {
			// Release already exists, return its ID
			return existingID, nil
		}
		if err != sql.ErrNoRows {
			// Some other error occurred
			return 0, err
		}
		// sql.ErrNoRows means release doesn't exist, proceed with insert
	}
	
	// Insert the new release
	query := `INSERT INTO releases (artist, title, year, musicbrainz_release_group_id) VALUES (?, ?, ?, ?)`
	result, err := db.Exec(query, artist, title, year, mbid)
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

func InsertOwnership(db *sql.DB, input OwnershipInput) (int64, error) {
	query := `INSERT INTO ownership (release_id, format_category, format_detail, acquired_date, cost, source, notes, discogs_release_id, is_promo, is_pirate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query, input.ReleaseID, input.FormatCategory, input.FormatDetail, input.AcquiredDate, input.Cost, input.Source, input.Notes, input.DiscogsID, input.IsPromo, input.IsPirate)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetReleaseID(db *sql.DB, mbid string) (int64, error) {
	var id int64
	err := db.QueryRow("SELECT id FROM releases WHERE musicbrainz_release_group_id = ?", mbid).Scan(&id)
	return id, err
}

func UpdateOwnership(db *sql.DB, ownershipID int64, update OwnershipUpdate) error {
	var setClauses []string
	var args []interface{}

	if update.FormatDetail != nil {
		setClauses = append(setClauses, "format_detail = ?")
		args = append(args, *update.FormatDetail)
	}
	if update.AcquiredDate != nil {
		setClauses = append(setClauses, "acquired_date = ?")
		args = append(args, *update.AcquiredDate)
	}
	if update.Cost != nil {
		setClauses = append(setClauses, "cost = ?")
		args = append(args, *update.Cost)
	}
	if update.Source != nil {
		setClauses = append(setClauses, "source = ?")
		args = append(args, *update.Source)
	}
	if update.Notes != nil {
		setClauses = append(setClauses, "notes = ?")
		args = append(args, *update.Notes)
	}
	if update.DiscogsID != nil {
		setClauses = append(setClauses, "discogs_release_id = ?")
		args = append(args, *update.DiscogsID)
	}
	if update.IsPromo != nil {
		setClauses = append(setClauses, "is_promo = ?")
		args = append(args, *update.IsPromo)
	}
	if update.IsPirate != nil {
		setClauses = append(setClauses, "is_pirate = ?")
		args = append(args, *update.IsPirate)
	}

	if len(setClauses) == 0 {
		return nil // Nothing to update
	}

	query := "UPDATE ownership SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	args = append(args, ownershipID)

	_, err := db.Exec(query, args...)
	return err
}

// GetOrCreateTag finds a tag by name or creates it if it doesn't exist.
// Returns the tag ID.
func GetOrCreateTag(db *sql.DB, name string) (int64, error) {
	// Try to find existing tag
	var id int64
	err := db.QueryRow("SELECT id FROM tags WHERE name = ?", name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	
	// Tag doesn't exist, create it
	result, err := db.Exec("INSERT INTO tags (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// AddTagToOwnership associates a tag with an ownership record.
// Silently ignores duplicates.
func AddTagToOwnership(db *sql.DB, ownershipID int64, tagID int64) error {
	_, err := db.Exec("INSERT OR IGNORE INTO ownership_tags (ownership_id, tag_id) VALUES (?, ?)", ownershipID, tagID)
	return err
}

// RemoveTagFromOwnership removes the association between a tag and ownership.
func RemoveTagFromOwnership(db *sql.DB, ownershipID int64, tagID int64) error {
	_, err := db.Exec("DELETE FROM ownership_tags WHERE ownership_id = ? AND tag_id = ?", ownershipID, tagID)
	return err
}

// RemoveAllTagsFromOwnership removes all tag associations from an ownership record.
func RemoveAllTagsFromOwnership(db *sql.DB, ownershipID int64) error {
	_, err := db.Exec("DELETE FROM ownership_tags WHERE ownership_id = ?", ownershipID)
	return err
}

// GetTagsForOwnership retrieves all tags for a given ownership record.
func GetTagsForOwnership(db *sql.DB, ownershipID int64) ([]string, error) {
	rows, err := db.Query(`
		SELECT t.name 
		FROM tags t
		JOIN ownership_tags ot ON t.id = ot.tag_id
		WHERE ot.ownership_id = ?
		ORDER BY t.name`, ownershipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, rows.Err()
}

// GetTagIDByName retrieves a tag ID by its name.
func GetTagIDByName(db *sql.DB, name string) (int64, error) {
	var id int64
	err := db.QueryRow("SELECT id FROM tags WHERE name = ?", name).Scan(&id)
	return id, err
}

// RenameTag updates a tag's name.
func RenameTag(db *sql.DB, oldName string, newName string) error {
	_, err := db.Exec("UPDATE tags SET name = ? WHERE name = ?", newName, oldName)
	return err
}

// DeleteTag removes a tag and all its associations.
func DeleteTag(db *sql.DB, name string) error {
	// First get the tag ID
	tagID, err := GetTagIDByName(db, name)
	if err != nil {
		return err
	}
	
	// Delete associations first
	_, err = db.Exec("DELETE FROM ownership_tags WHERE tag_id = ?", tagID)
	if err != nil {
		return err
	}
	
	// Delete the tag
	_, err = db.Exec("DELETE FROM tags WHERE id = ?", tagID)
	return err
}
