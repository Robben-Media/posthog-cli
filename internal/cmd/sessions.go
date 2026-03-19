package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/builtbyrobben/posthog-cli/internal/outfmt"
)

type SessionsCmd struct {
	List SessionsListCmd `cmd:"" help:"List session recordings"`
	Get  SessionsGetCmd  `cmd:"" help:"Get session recording details"`
}

type SessionsListCmd struct {
	Days  int `help:"Number of days to look back" default:"7"`
	Limit int `help:"Maximum number of results" default:"20"`
}

func (cmd *SessionsListCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	dateFrom := time.Now().AddDate(0, 0, -cmd.Days).Format("2006-01-02T15:04:05")
	path := fmt.Sprintf("/api/projects/%s/session_recordings/?date_from=%s&limit=%d",
		client.ProjectID(), dateFrom, cmd.Limit)

	var result struct {
		Results []json.RawMessage `json:"results"`
	}

	if err := client.Get(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result.Results)
	}

	type session struct {
		ID                 string  `json:"id"`
		StartTime          string  `json:"start_time"`
		EndTime            string  `json:"end_time"`
		RecordingDuration  float64 `json:"recording_duration"`
		ClickCount         float64 `json:"click_count"`
		KeypressCount      float64 `json:"keypress_count"`
		ActiveSeconds      float64 `json:"active_seconds"`
		DistinctID         string  `json:"distinct_id"`
		StartURL           string  `json:"start_url"`
	}

	var sessions []session

	for _, raw := range result.Results {
		var s session
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}

		sessions = append(sessions, s)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "START_TIME", "DURATION_S", "CLICKS", "START_URL"}
		var rows [][]string

		for _, s := range sessions {
			rows = append(rows, []string{
				s.ID,
				s.StartTime,
				fmt.Sprintf("%.0f", s.RecordingDuration),
				fmt.Sprintf("%.0f", s.ClickCount),
				s.StartURL,
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(sessions) == 0 {
		fmt.Fprintln(os.Stderr, "No session recordings found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d session(s)\n\n", len(sessions))

	for _, s := range sessions {
		fmt.Printf("ID: %s\n", s.ID)
		fmt.Printf("  Start:    %s\n", s.StartTime)
		fmt.Printf("  Duration: %.0fs\n", s.RecordingDuration)
		fmt.Printf("  Clicks:   %.0f\n", s.ClickCount)

		if s.StartURL != "" {
			fmt.Printf("  URL:      %s\n", s.StartURL)
		}

		if s.DistinctID != "" {
			fmt.Printf("  Person:   %s\n", s.DistinctID)
		}

		fmt.Println()
	}

	return nil
}

type SessionsGetCmd struct {
	ID string `arg:"" required:"" help:"Session recording ID"`
}

func (cmd *SessionsGetCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/projects/%s/session_recordings/%s/", client.ProjectID(), cmd.ID)

	var result map[string]any
	if err := client.Get(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	id, _ := result["id"].(string)
	startTime, _ := result["start_time"].(string)
	endTime, _ := result["end_time"].(string)
	duration, _ := result["recording_duration"].(float64)
	clicks, _ := result["click_count"].(float64)
	keys, _ := result["keypress_count"].(float64)
	distinctID, _ := result["distinct_id"].(string)
	startURL, _ := result["start_url"].(string)

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "START_TIME", "END_TIME", "DURATION_S", "CLICKS", "KEYS", "PERSON", "URL"}
		rows := [][]string{{id, startTime, endTime, fmt.Sprintf("%.0f", duration),
			fmt.Sprintf("%.0f", clicks), fmt.Sprintf("%.0f", keys), distinctID, startURL}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID:         %s\n", id)
	fmt.Printf("Start:      %s\n", startTime)
	fmt.Printf("End:        %s\n", endTime)
	fmt.Printf("Duration:   %.0fs\n", duration)
	fmt.Printf("Clicks:     %.0f\n", clicks)
	fmt.Printf("Keypresses: %.0f\n", keys)

	if distinctID != "" {
		fmt.Printf("Person:     %s\n", distinctID)
	}

	if startURL != "" {
		fmt.Printf("Start URL:  %s\n", startURL)
	}

	// Show the watch URL
	baseURL := client.BaseURL()
	fmt.Printf("\nWatch: %s/replay/%s\n", baseURL, id)

	return nil
}
