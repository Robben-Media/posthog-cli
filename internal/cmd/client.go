package cmd

import (
	"fmt"
	"os"

	"github.com/Robben-Media/posthog-cli/internal/api"
	"github.com/Robben-Media/posthog-cli/internal/config"
	"github.com/Robben-Media/posthog-cli/internal/secrets"
)

// getPostHogClient creates an API client from the keyring and config file.
func getPostHogClient() (*api.Client, error) {
	// Check for environment variable override first
	apiKey := os.Getenv("POSTHOG_API_KEY")

	if apiKey == "" {
		store, err := secrets.OpenDefault()
		if err != nil {
			return nil, fmt.Errorf("open credential store: %w", err)
		}

		apiKey, err = store.GetAPIKey()
		if err != nil {
			return nil, fmt.Errorf("get API key: %w (set POSTHOG_API_KEY or run 'posthog-cli auth set-key')", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if cfg.ProjectID == "" {
		// Check env var fallback
		cfg.ProjectID = os.Getenv("POSTHOG_PROJECT_ID")
	}

	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("project ID not configured; run 'posthog-cli auth set-key' and set project_id in config")
	}

	return api.NewClient(apiKey,
		api.WithBaseURL(cfg.BaseURL()),
		api.WithProjectID(cfg.ProjectID),
		api.WithUserAgent("posthog-cli/"+VersionString()),
	), nil
}
