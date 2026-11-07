package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Config holds the posthog-cli configuration (project ID and region).
// The API key is stored in the keyring, not here.
type Config struct {
	ProjectID string `json:"project_id"`
	Region    string `json:"region"` // "us" or "eu"
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Region: "us",
	}
}

// Load reads the config file from disk. Returns default config if file doesn't exist.
func Load() (*Config, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, fmt.Errorf("config path: %w", err)
	}

	// Anchor the fixed config name to the user-selected config directory.
	// Root prevents a config symlink from escaping that directory.
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}

		return nil, fmt.Errorf("resolve config directory: %w", err)
	}

	root, err := os.OpenRoot(resolvedDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}

		return nil, fmt.Errorf("open config directory: %w", err)
	}
	defer root.Close()

	resolvedPath, err := filepath.EvalSymlinks(filepath.Join(resolvedDir, "config.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}

		return nil, fmt.Errorf("resolve config: %w", err)
	}

	relativePath, err := filepath.Rel(resolvedDir, resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("relative config path: %w", err)
	}

	f, err := root.Open(relativePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}

		return nil, fmt.Errorf("read config: %w", err)
	}

	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	dir, err := EnsureConfigDir()
	if err != nil {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	path := dir + "/config.json"

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// BaseURL returns the PostHog API base URL for the configured region.
func (c *Config) BaseURL() string {
	switch c.Region {
	case "eu":
		return "https://eu.posthog.com"
	default:
		return "https://us.posthog.com"
	}
}
