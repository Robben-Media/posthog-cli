package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/Robben-Media/posthog-cli/internal/outfmt"
)

type EventsCmd struct {
	List EventsListCmd `cmd:"" help:"List events with optional filters"`
}

type EventsListCmd struct {
	Event    string `help:"Filter by event name (e.g. '$pageview')" name:"event"`
	Property string `help:"Filter by property (e.g. '$current_url=/roofing/')" name:"property"`
	Days     int    `help:"Number of days to look back" default:"30"`
	Limit    int    `help:"Maximum number of results" default:"50"`
}

func (cmd *EventsListCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	// Build HogQL query for events
	where := "1=1"

	if cmd.Event != "" {
		where += fmt.Sprintf(" AND event = '%s'", escapeHogQL(cmd.Event))
	}

	if cmd.Property != "" {
		// Parse property=value format
		key, value := parsePropertyFilter(cmd.Property)
		if key != "" && value != "" {
			where += fmt.Sprintf(" AND properties.%s = '%s'", escapeHogQL(key), escapeHogQL(value))
		}
	}

	hogql := fmt.Sprintf(
		"SELECT event, properties.$current_url as url, person.properties.email as person, "+
			"timestamp FROM events WHERE %s AND timestamp > now() - interval %d day "+
			"ORDER BY timestamp DESC LIMIT %d",
		where, cmd.Days, cmd.Limit,
	)

	data, err := client.Query(ctx, hogql)
	if err != nil {
		return fmt.Errorf("query events: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return writeRawJSON(os.Stdout, data)
	}

	return formatQueryResults(ctx, data)
}

// parsePropertyFilter splits "$key=value" or "$key=/value/" into key and value.
func parsePropertyFilter(s string) (string, string) {
	for i, c := range s {
		if c == '=' {
			return s[:i], s[i+1:]
		}
	}

	return s, ""
}

func escapeHogQL(s string) string {
	// Simple single-quote escape for HogQL string literals
	result := make([]byte, 0, len(s))
	for i := range len(s) {
		if s[i] == '\'' {
			result = append(result, '\\', '\'')
		} else {
			result = append(result, s[i])
		}
	}

	return string(result)
}

// writeRawJSON writes raw JSON bytes formatted to the writer.
func writeRawJSON(w *os.File, data []byte) error {
	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		// Not valid JSON, write as-is
		_, err = w.Write(data)
		return err
	}

	return outfmt.WriteJSON(w, parsed)
}

// formatQueryResults prints HogQL query results as a human-readable table.
func formatQueryResults(ctx context.Context, data []byte) error {
	var result struct {
		Columns []string `json:"columns"`
		Results [][]any  `json:"results"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("parse query results: %w", err)
	}

	if outfmt.IsPlain(ctx) {
		var rows [][]string

		for _, row := range result.Results {
			var cells []string
			for _, cell := range row {
				cells = append(cells, fmt.Sprintf("%v", cell))
			}

			rows = append(rows, cells)
		}

		return outfmt.WritePlain(os.Stdout, result.Columns, rows)
	}

	// Human-readable colored output
	if len(result.Results) == 0 {
		fmt.Fprintln(os.Stderr, "No results")
		return nil
	}

	fmt.Fprintf(os.Stderr, "%d result(s)\n\n", len(result.Results))

	// Print header
	for i, col := range result.Columns {
		if i > 0 {
			fmt.Print("  |  ")
		}

		fmt.Print(col)
	}

	fmt.Println()

	// Separator
	for range result.Columns {
		fmt.Print("----------")
	}

	fmt.Println()

	// Print rows
	for _, row := range result.Results {
		for i, cell := range row {
			if i > 0 {
				fmt.Print("  |  ")
			}

			fmt.Printf("%v", cell)
		}

		fmt.Println()
	}

	return nil
}

// buildListURL creates a paginated list URL with query parameters.
func buildListURL(basePath string, params url.Values) string {
	if len(params) == 0 {
		return basePath
	}

	return basePath + "?" + params.Encode()
}
