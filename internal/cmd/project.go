package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/builtbyrobben/posthog-cli/internal/outfmt"
)

type ProjectCmd struct {
	Info ProjectInfoCmd `cmd:"" help:"Show project name, timezone, settings"`
	List ProjectListCmd `cmd:"" help:"List all projects in the organization"`
}

type ProjectInfoCmd struct{}

func (cmd *ProjectInfoCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/projects/%s/", client.ProjectID())

	var result map[string]any
	if err := client.Get(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("get project info: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	name, _ := result["name"].(string)
	timezone, _ := result["timezone"].(string)
	id, _ := result["id"].(float64)

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TIMEZONE"}
		rows := [][]string{{fmt.Sprintf("%.0f", id), name, timezone}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID:       %.0f\n", id)
	fmt.Printf("Name:     %s\n", name)
	fmt.Printf("Timezone: %s\n", timezone)

	if completed, ok := result["completed_snippet_onboarding"].(bool); ok {
		fmt.Printf("Onboarded: %t\n", completed)
	}

	return nil
}

type ProjectListCmd struct{}

func (cmd *ProjectListCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	var result struct {
		Results []json.RawMessage `json:"results"`
	}

	if err := client.Get(ctx, "/api/projects/", nil, &result); err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result.Results)
	}

	type project struct {
		ID       float64 `json:"id"`
		Name     string  `json:"name"`
		Timezone string  `json:"timezone"`
	}

	var projects []project

	for _, raw := range result.Results {
		var p project
		if err := json.Unmarshal(raw, &p); err != nil {
			continue
		}

		projects = append(projects, p)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TIMEZONE"}
		var rows [][]string

		for _, p := range projects {
			rows = append(rows, []string{fmt.Sprintf("%.0f", p.ID), p.Name, p.Timezone})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "No projects found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d project(s)\n\n", len(projects))

	for _, p := range projects {
		fmt.Printf("ID: %.0f\n", p.ID)
		fmt.Printf("  Name:     %s\n", p.Name)
		fmt.Printf("  Timezone: %s\n", p.Timezone)
		fmt.Println()
	}

	return nil
}
