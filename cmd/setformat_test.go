package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"musiccat/internal/config"
)

func TestSetFormatCommand(t *testing.T) {
	tempDir := t.TempDir()
	tempConfig := filepath.Join(tempDir, "config.toml")
	
	originalGetConfigPath := config.GetConfigPath
	config.GetConfigPath = func() (string, error) {
		return tempConfig, nil
	}
	defer func() {
		config.GetConfigPath = originalGetConfigPath
	}()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantFmt string
	}{
		{"valid lowercase cd", []string{"cd"}, false, "CD"},
		{"valid uppercase CD", []string{"CD"}, false, "CD"},
		{"valid mixed Vinyl", []string{"Vinyl"}, false, "Vinyl"},
		{"valid cassette", []string{"cassette"}, false, "Cassette"},
		{"invalid format", []string{"digital"}, true, ""},
		{"no args", []string{}, true, ""},
		{"too many args", []string{"cd", "vinyl"}, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(tempConfig)
			
			setFormatCmd.SetArgs(tt.args)
			err := setFormatCmd.Execute()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("setFormatCmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				cfg, err := config.LoadConfig()
				if err != nil {
					t.Fatalf("LoadConfig() error = %v", err)
				}
				if cfg.CurrentFormat != tt.wantFmt {
					t.Errorf("CurrentFormat = %q, want %q", cfg.CurrentFormat, tt.wantFmt)
				}
			}
		})
	}
}
