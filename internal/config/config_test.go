package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path, err := GetConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows
	}
	expected := filepath.Join(home, ".musiccat", "config.toml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestLoadConfig(t *testing.T) {
	path, _ := GetConfigPath()
	os.Remove(path) // Clean up from previous tests
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.CurrentFormat != "" {
		t.Errorf("expected empty, got %s", cfg.CurrentFormat)
	}
}

func TestSaveLoadConfig(t *testing.T) {
	cfg := &Config{CurrentFormat: "CD"}
	err := SaveConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.CurrentFormat != "CD" {
		t.Errorf("expected CD, got %s", loaded.CurrentFormat)
	}
	// Reset for other tests
	err = SaveConfig(&Config{})
	if err != nil {
		t.Fatal(err)
	}
}
