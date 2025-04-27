package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type (
	Server struct {
		Host string `env:"host"`
		Port int    `env:"port"`
	}

	DB struct {
		Path string `env:"path"`
	}

	Config struct {
		Server      Server      `env:"server"`
		Environment Environment `env:"environment"`
		DB          DB          `env:"db"`
	}
)

// SetDefaults sets default values for configuration fields if they are not already set.
// Specifically, it sets a default database path if `DB.Path` is empty.
func (c *Config) SetDefaults() error {
	if c.DB.Path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		defaultDBPath := filepath.Join(homeDir, ".cardmax", "data.db")
		c.DB.Path = defaultDBPath
	}
	return nil
}
