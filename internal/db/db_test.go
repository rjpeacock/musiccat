package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestGetDBPath(t *testing.T) {
	path, err := GetDBPath()
	if err != nil {
		t.Fatal(err)
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}
	expected := filepath.Join(home, ".musiccat", "musiccat.db")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestOpenDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB test in sandbox")
	}
	db := openTestDB(t)
	defer db.Close()
	if db == nil {
		t.Fatal("db is nil")
	}
}

func TestBootstrapDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB test in sandbox")
	}
	db := openTestDB(t)
	defer db.Close()
	err := BootstrapDB(db)
	if err != nil {
		t.Fatal(err)
	}
	// Check tables exist
	_, err = db.Exec("SELECT 1 FROM releases LIMIT 1")
	if err != nil {
		t.Fatal("releases table not created")
	}
	_, err = db.Exec("SELECT 1 FROM ownership LIMIT 1")
	if err != nil {
		t.Fatal("ownership table not created")
	}
}

func ptr[T any](v T) *T { return &v }

func insertRelease(t *testing.T, db *sql.DB, artist, title string, year *int, mbid *string) (int64, error) {
	query := `INSERT INTO releases (artist, title, year, musicbrainz_release_group_id) VALUES (?, ?, ?, ?)`
	var args []interface{}
	args = append(args, artist, title)
	if year != nil {
		args = append(args, *year)
	} else {
		args = append(args, nil)
	}
	if mbid != nil {
		args = append(args, *mbid)
	} else {
		args = append(args, nil)
	}
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func insertOwnership(t *testing.T, db *sql.DB, releaseID int64, formatCategory string, formatDetail *string, purchaseDate *string, cost *float64, source *string, notes *string, discogsID *int) (int64, error) {
	query := `INSERT INTO ownership (release_id, format_category, format_detail, purchase_date, cost, source, notes, discogs_release_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	var args []interface{}
	args = append(args, releaseID, formatCategory)
	if formatDetail != nil {
		args = append(args, *formatDetail)
	} else {
		args = append(args, nil)
	}
	if purchaseDate != nil {
		args = append(args, *purchaseDate)
	} else {
		args = append(args, nil)
	}
	if cost != nil {
		args = append(args, *cost)
	} else {
		args = append(args, nil)
	}
	if source != nil {
		args = append(args, *source)
	} else {
		args = append(args, nil)
	}
	if notes != nil {
		args = append(args, *notes)
	} else {
		args = append(args, nil)
	}
	if discogsID != nil {
		args = append(args, *discogsID)
	} else {
		args = append(args, nil)
	}
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func TestInsertRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB test in sandbox")
	}
	db := openTestDB(t)
	defer db.Close()
	BootstrapDB(db)
	tests := []struct {
		name    string
		artist  string
		title   string
		year    *int
		mbid    *string
		wantErr bool
	}{
		{"valid", "Artist1", "Title1", ptr(2020), ptr("mbid1"), false},
		{"no year", "Artist2", "Title2", nil, ptr("mbid2"), false},
		{"no mbid", "Artist3", "Title3", ptr(2021), nil, false},
		{"duplicate mbid", "Artist4", "Title4", nil, ptr("mbid1"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := insertRelease(t, db, tt.artist, tt.title, tt.year, tt.mbid)
			if (err != nil) != tt.wantErr {
				t.Errorf("insertRelease() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInsertOwnership(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping DB test in sandbox")
	}
	db := openTestDB(t)
	defer db.Close()
	BootstrapDB(db)
	// Insert a release first
	releaseID, err := insertRelease(t, db, "Artist", "Title", ptr(2020), ptr("mbid"))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name           string
		releaseID      int64
		formatCategory string
		wantErr        bool
	}{
		{"valid", releaseID, "CD", false},
		{"invalid release id", 999, "Vinyl", true}, // foreign key constraint
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := insertOwnership(t, db, tt.releaseID, tt.formatCategory, nil, nil, nil, nil, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("insertOwnership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
