package core

import (
	"os"
)

// Config holds the telemetry configuration options.
// Telemetry v2 — simplified, no SQLite, no plugins.
type Config struct {
	ProjectPath string `toml:"project_path"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ProjectPath: ".",
	}
}

// LoadConfig detects the MVGN project root and returns a Config.
// In v2, the config is minimal — just the project path auto-detected.
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	// Try to detect an MVGN project root
	if _, err := os.Stat(".mvgn-context.json"); err == nil {
		cfg.ProjectPath = "."
	} else if _, err := os.Stat("../.mvgn-context.json"); err == nil {
		cfg.ProjectPath = ".."
	}

	return cfg, nil
}
