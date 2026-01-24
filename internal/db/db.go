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

	return sql.Open("sqlite", path)
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
    purchase_date TEXT,
    cost REAL,
    source TEXT,
    notes TEXT,
    discogs_release_id INTEGER,
    is_promo BOOLEAN DEFAULT FALSE
);
`

func BootstrapDB(db *sql.DB) error {
	if _, err := db.Exec(createReleasesTable); err != nil {
		return err
	}
	if _, err := db.Exec(createOwnershipTable); err != nil {
		return err
	}
	// Migration: add is_promo column if missing
	_, err := db.Exec("ALTER TABLE ownership ADD COLUMN is_promo BOOLEAN DEFAULT FALSE")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}

func InsertRelease(db *sql.DB, artist, title string, year *int, mbid *string) (int64, error) {
	query := `INSERT OR IGNORE INTO releases (artist, title, year, musicbrainz_release_group_id) VALUES (?, ?, ?, ?)`
	result, err := db.Exec(query, artist, title, year, mbid)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func InsertOwnership(db *sql.DB, releaseID int64, formatCategory string, formatDetail *string, purchaseDate *string, cost *float64, source *string, notes *string, discogsID *int, isPromo bool) (int64, error) {
	query := `INSERT INTO ownership (release_id, format_category, format_detail, purchase_date, cost, source, notes, discogs_release_id, is_promo) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query, releaseID, formatCategory, formatDetail, purchaseDate, cost, source, notes, discogsID, isPromo)
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
