package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/builtbyrobben/posthog-cli/internal/outfmt"
)

type DashboardCmd struct {
	List DashboardListCmd `cmd:"" help:"List all dashboards"`
	Get  DashboardGetCmd  `cmd:"" help:"Get dashboard with insight summaries"`
}

type DashboardListCmd struct {
	Limit int `help:"Maximum number of results" default:"50"`
}

func (cmd *DashboardListCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/projects/%s/dashboards/?limit=%d", client.ProjectID(), cmd.Limit)

	var result struct {
		Results []json.RawMessage `json:"results"`
	}

	if err := client.Get(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("list dashboards: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result.Results)
	}

	type dashboard struct {
		ID          float64 `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Pinned      bool    `json:"pinned"`
		CreatedAt   string  `json:"created_at"`
	}

	var dashboards []dashboard

	for _, raw := range result.Results {
		var d dashboard
		if err := json.Unmarshal(raw, &d); err != nil {
			continue
		}

		dashboards = append(dashboards, d)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PINNED", "DESCRIPTION"}
		var rows [][]string

		for _, d := range dashboards {
			rows = append(rows, []string{
				fmt.Sprintf("%.0f", d.ID),
				d.Name,
				fmt.Sprintf("%t", d.Pinned),
				d.Description,
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(dashboards) == 0 {
		fmt.Fprintln(os.Stderr, "No dashboards found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d dashboard(s)\n\n", len(dashboards))

	for _, d := range dashboards {
		pinned := ""
		if d.Pinned {
			pinned = " [pinned]"
		}

		fmt.Printf("ID: %.0f%s\n", d.ID, pinned)
		fmt.Printf("  Name: %s\n", d.Name)

		if d.Description != "" {
			fmt.Printf("  Desc: %s\n", d.Description)
		}

		fmt.Println()
	}

	return nil
}

type DashboardGetCmd struct {
	ID string `arg:"" required:"" help:"Dashboard ID"`
}

func (cmd *DashboardGetCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/projects/%s/dashboards/%s/", client.ProjectID(), cmd.ID)

	var result map[string]any
	if err := client.Get(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("get dashboard: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	name, _ := result["name"].(string)
	description, _ := result["description"].(string)
	id, _ := result["id"].(float64)
	pinned, _ := result["pinned"].(bool)

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PINNED", "DESCRIPTION"}
		rows := [][]string{{fmt.Sprintf("%.0f", id), name, fmt.Sprintf("%t", pinned), description}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID:       %.0f\n", id)
	fmt.Printf("Name:     %s\n", name)
	fmt.Printf("Pinned:   %t\n", pinned)

	if description != "" {
		fmt.Printf("Desc:     %s\n", description)
	}

	// List tiles/insights
	if tiles, ok := result["tiles"].([]any); ok && len(tiles) > 0 {
		fmt.Printf("\nInsights (%d):\n", len(tiles))

		for i, tile := range tiles {
			if t, ok := tile.(map[string]any); ok {
				if insight, ok := t["insight"].(map[string]any); ok {
					insightName, _ := insight["name"].(string)
					insightID, _ := insight["id"].(float64)

					if insightName == "" {
						insightName = "(unnamed)"
					}

					fmt.Printf("  %d. [%.0f] %s\n", i+1, insightID, insightName)
				}
			}
		}
	}

	return nil
}
