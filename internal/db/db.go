package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
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

	return sql.Open("sqlite3", path)
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
    release_id INTEGER NOT NULL,
    format_category TEXT NOT NULL,
    format_detail TEXT,
    purchase_date TEXT,
    cost REAL,
    source TEXT,
    notes TEXT,
    discogs_release_id INTEGER
);
`

func BootstrapDB(db *sql.DB) error {
	if _, err := db.Exec(createReleasesTable); err != nil {
		return err
	}
	if _, err := db.Exec(createOwnershipTable); err != nil {
		return err
	}
	return nil
}
