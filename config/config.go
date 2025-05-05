package config

import (
	"fmt"
	"net/http"
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

	Session struct {
		Secret   string `env:"secret"`
		MaxAge   int    `env:"max_age"`
		Secure   bool   `env:"secure"`
		SameSite string `env:"same_site"`
	}

	Config struct {
		Server      Server      `env:"server"`
		Environment Environment `env:"environment"`
		DB          DB          `env:"db"`
		Session     Session     `env:"session"`
	}
)

// isValidSameSite checks if the provided SameSite value is valid
func isValidSameSite(value string) bool {
	switch value {
	case "none", "lax", "strict":
		return true
	default:
		return false
	}
}

// GetSameSiteMode converts a string SameSite value to http.SameSite
func (s Session) GetSameSiteMode() http.SameSite {
	switch s.SameSite {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode // Default to Lax if invalid
	}
}

// SetDefaults sets default values for configuration fields if they are not already set.
// Specifically, it sets defaults for database path and session settings.
func (c *Config) SetDefaults() error {
	// Set default DB path
	if c.DB.Path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		defaultDBPath := filepath.Join(homeDir, ".cardmax", "data.db")
		c.DB.Path = defaultDBPath
	}

	// Set default session settings
	if c.Session.Secret == "" {
		// This is just a fallback, in production the secret should be set in environment variables
		c.Session.Secret = "cardmax-default-secret-key-replace-me-in-production"
	}

	if c.Session.MaxAge <= 0 {
		// Default session age: 7 days
		c.Session.MaxAge = 86400 * 7
	}

	if c.Session.SameSite == "" {
		c.Session.SameSite = "lax" // Default to SameSite=Lax for good security/usability balance
	}

	// Validate session settings
	if !isValidSameSite(c.Session.SameSite) {
		return fmt.Errorf("invalid SameSite value: %s (valid values: none, lax, strict)", c.Session.SameSite)
	}

	return nil
}
