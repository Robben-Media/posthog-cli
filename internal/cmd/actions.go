package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/builtbyrobben/posthog-cli/internal/outfmt"
)

type ActionsCmd struct {
	List ActionsListCmd `cmd:"" help:"List all actions"`
	Get  ActionsGetCmd  `cmd:"" help:"Get action details by ID"`
}

type ActionsListCmd struct {
	Limit int `help:"Maximum number of results" default:"50"`
}

func (cmd *ActionsListCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/projects/%s/actions/?limit=%d", client.ProjectID(), cmd.Limit)

	var result struct {
		Results []json.RawMessage `json:"results"`
	}

	if err := client.Get(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("list actions: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result.Results)
	}

	type action struct {
		ID          float64 `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		CreatedAt   string  `json:"created_at"`
	}

	var actions []action

	for _, raw := range result.Results {
		var a action
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}

		actions = append(actions, a)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "DESCRIPTION"}
		var rows [][]string

		for _, a := range actions {
			rows = append(rows, []string{fmt.Sprintf("%.0f", a.ID), a.Name, a.Description})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(actions) == 0 {
		fmt.Fprintln(os.Stderr, "No actions found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d action(s)\n\n", len(actions))

	for _, a := range actions {
		fmt.Printf("ID: %.0f\n", a.ID)
		fmt.Printf("  Name:        %s\n", a.Name)

		if a.Description != "" {
			fmt.Printf("  Description: %s\n", a.Description)
		}

		fmt.Println()
	}

	return nil
}

type ActionsGetCmd struct {
	ID string `arg:"" required:"" help:"Action ID"`
}

func (cmd *ActionsGetCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/projects/%s/actions/%s/", client.ProjectID(), cmd.ID)

	var result map[string]any
	if err := client.Get(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("get action: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	name, _ := result["name"].(string)
	description, _ := result["description"].(string)
	id, _ := result["id"].(float64)
	createdAt, _ := result["created_at"].(string)

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "DESCRIPTION", "CREATED_AT"}
		rows := [][]string{{fmt.Sprintf("%.0f", id), name, description, createdAt}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID:          %.0f\n", id)
	fmt.Printf("Name:        %s\n", name)

	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}

	if createdAt != "" {
		fmt.Printf("Created:     %s\n", createdAt)
	}

	// Print steps if available
	if steps, ok := result["steps"].([]any); ok && len(steps) > 0 {
		fmt.Printf("\nSteps (%d):\n", len(steps))

		for i, step := range steps {
			if s, ok := step.(map[string]any); ok {
				event, _ := s["event"].(string)
				tag, _ := s["tag_name"].(string)
				fmt.Printf("  %d. event=%s", i+1, event)

				if tag != "" {
					fmt.Printf(" tag=%s", tag)
				}

				fmt.Println()
			}
		}
	}

	return nil
}
