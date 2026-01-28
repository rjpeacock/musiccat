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

func insertOwnership(t *testing.T, db *sql.DB, releaseID int64, formatCategory string, formatDetail *string, acquiredDate *string, cost *float64, source *string, notes *string, discogsID *int, isPromo bool, isPirate bool) (int64, error) {
	query := `INSERT INTO ownership (release_id, format_category, format_detail, acquired_date, cost, source, notes, discogs_release_id, is_promo, is_pirate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	var args []interface{}
	args = append(args, releaseID, formatCategory)
	if formatDetail != nil {
		args = append(args, *formatDetail)
	} else {
		args = append(args, nil)
	}
	if acquiredDate != nil {
		args = append(args, *acquiredDate)
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
	args = append(args, isPromo, isPirate)
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
			_, err := insertOwnership(t, db, tt.releaseID, tt.formatCategory, nil, nil, nil, nil, nil, nil, false, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("insertOwnership() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInsertOwnershipPromo(t *testing.T) {
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
	// Insert ownership with promo false
	id1, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	// Insert with promo true
	id2, err := insertOwnership(t, db, releaseID, "Vinyl", nil, nil, nil, nil, nil, nil, true, false)
	if err != nil {
		t.Fatal(err)
	}
	// Check DB
	var promo1, promo2 bool
	err = db.QueryRow("SELECT is_promo FROM ownership WHERE id = ?", id1).Scan(&promo1)
	if err != nil {
		t.Fatal(err)
	}
	err = db.QueryRow("SELECT is_promo FROM ownership WHERE id = ?", id2).Scan(&promo2)
	if err != nil {
		t.Fatal(err)
	}
	if promo1 != false {
		t.Errorf("expected promo false, got %v", promo1)
	}
	if promo2 != true {
		t.Errorf("expected promo true, got %v", promo2)
	}
}

func TestInsertOwnershipPirate(t *testing.T) {
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
	// Insert ownership with pirate false
	id1, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	// Insert with pirate true
	id2, err := insertOwnership(t, db, releaseID, "Vinyl", nil, nil, nil, nil, nil, nil, false, true)
	if err != nil {
		t.Fatal(err)
	}
	// Check DB
	var pirate1, pirate2 bool
	err = db.QueryRow("SELECT is_pirate FROM ownership WHERE id = ?", id1).Scan(&pirate1)
	if err != nil {
		t.Fatal(err)
	}
	err = db.QueryRow("SELECT is_pirate FROM ownership WHERE id = ?", id2).Scan(&pirate2)
	if err != nil {
		t.Fatal(err)
	}
	if pirate1 != false {
		t.Errorf("expected pirate false, got %v", pirate1)
	}
	if pirate2 != true {
		t.Errorf("expected pirate true, got %v", pirate2)
	}
}

func TestUpdateOwnership(t *testing.T) {
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

	// Insert ownership
	id, err := insertOwnership(t, db, releaseID, "CD", nil, nil, nil, nil, nil, nil, false, false)
	if err != nil {
		t.Fatal(err)
	}

	// Test update with various fields
	newFormatDetail := "CD Single"
	newAcquiredDate := "2023-01-01"
	newCost := 15.99
	newSource := "Test Store"
	newNotes := "Test notes"
	newPromo := true
	newPirate := false

	err = UpdateOwnership(db, id, &newFormatDetail, &newAcquiredDate, &newCost, &newSource, &newNotes, nil, &newPromo, &newPirate)
	if err != nil {
		t.Fatal(err)
	}

	// Verify updates
	var updatedFormatDetail, updatedAcquiredDate string
	var updatedCost float64
	var updatedSource, updatedNotes *string
	var updatedPromo, updatedPirate bool

	err = db.QueryRow(`
		SELECT format_detail, acquired_date, cost, source, notes, is_promo, is_pirate 
		FROM ownership WHERE id = ?`, id).Scan(
		&updatedFormatDetail, &updatedAcquiredDate, &updatedCost, &updatedSource, &updatedNotes, &updatedPromo, &updatedPirate)
	if err != nil {
		t.Fatal(err)
	}

	if updatedFormatDetail != newFormatDetail {
		t.Errorf("expected format_detail %s, got %s", newFormatDetail, updatedFormatDetail)
	}
	if updatedAcquiredDate != newAcquiredDate {
		t.Errorf("expected acquired_date %s, got %s", newAcquiredDate, updatedAcquiredDate)
	}
	if updatedCost != newCost {
		t.Errorf("expected cost %f, got %f", newCost, updatedCost)
	}
	if updatedSource == nil || *updatedSource != newSource {
		t.Errorf("expected source %s, got %v", newSource, updatedSource)
	}
	if updatedNotes == nil || *updatedNotes != newNotes {
		t.Errorf("expected notes %s, got %v", newNotes, updatedNotes)
	}
	if updatedPromo != newPromo {
		t.Errorf("expected promo %t, got %t", newPromo, updatedPromo)
	}
	if updatedPirate != newPirate {
		t.Errorf("expected pirate %t, got %t", newPirate, updatedPirate)
	}
}
