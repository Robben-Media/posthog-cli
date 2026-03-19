package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/posthog-cli/internal/outfmt"
)

type HeatmapCmd struct {
	URL  string `required:"" help:"URL path to get heatmap data for (e.g. /roofing/)"`
	Days int    `help:"Number of days to look back" default:"7"`
	Type string `help:"Heatmap type: click or scroll" default:"click" enum:"click,scroll"`
}

func (cmd *HeatmapCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	var hogql string

	switch cmd.Type {
	case "scroll":
		hogql = fmt.Sprintf(
			"SELECT "+
				"properties.$viewport_height as vh, "+
				"properties.$scroll_depth as scroll_depth, "+
				"count() as events "+
				"FROM events "+
				"WHERE event = '$pageview' "+
				"AND properties.$current_url LIKE '%%%s%%' "+
				"AND timestamp > now() - interval %d day "+
				"GROUP BY vh, scroll_depth "+
				"ORDER BY events DESC "+
				"LIMIT 500",
			escapeHogQL(cmd.URL), cmd.Days,
		)
	default: // click
		hogql = fmt.Sprintf(
			"SELECT "+
				"properties.$mouse_x as x, "+
				"properties.$mouse_y as y, "+
				"properties.$viewport_width as vw, "+
				"properties.$viewport_height as vh, "+
				"count() as clicks "+
				"FROM events "+
				"WHERE event = '$autocapture' "+
				"AND properties.$current_url LIKE '%%%s%%' "+
				"AND timestamp > now() - interval %d day "+
				"GROUP BY x, y, vw, vh "+
				"ORDER BY clicks DESC "+
				"LIMIT 500",
			escapeHogQL(cmd.URL), cmd.Days,
		)
	}

	data, err := client.Query(ctx, hogql)
	if err != nil {
		return fmt.Errorf("query heatmap: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return writeRawJSON(os.Stdout, data)
	}

	return formatQueryResults(ctx, data)
}
