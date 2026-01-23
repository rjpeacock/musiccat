package db

import (
	"os"
	"path/filepath"
	"testing"
)

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
	db, err := OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if db == nil {
		t.Fatal("db is nil")
	}
}

func TestBootstrapDB(t *testing.T) {
	db, err := OpenDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = BootstrapDB(db)
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
