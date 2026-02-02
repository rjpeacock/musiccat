package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	CurrentFormat string `toml:"current_format"`
}

const configDir = ".musiccat"
const configFile = "config.toml"

// GetConfigPath is a variable holding a function, allowing it to be mocked in tests
var GetConfigPath = func() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	var config Config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &config, nil
	}

	_, err = toml.DecodeFile(path, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(config)
}
