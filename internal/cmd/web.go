package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Robben-Media/posthog-cli/internal/outfmt"
)

type WebCmd struct {
	TopPages    WebTopPagesCmd    `cmd:"" help:"Top pages by pageviews"`
	ScrollDepth WebScrollDepthCmd `cmd:"" help:"Average scroll depth per URL"`
	Engagement  WebEngagementCmd  `cmd:"" help:"Rage clicks, dead clicks, quick backs"`
	Sources     WebSourcesCmd     `cmd:"" help:"Traffic sources"`
}

type WebTopPagesCmd struct {
	Days  int `help:"Number of days to look back" default:"30"`
	Limit int `help:"Maximum number of results" default:"20"`
}

func (cmd *WebTopPagesCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	hogql := fmt.Sprintf(
		"SELECT "+
			"properties.$current_url as url, "+
			"count() as views, "+
			"count(DISTINCT person_id) as unique_visitors "+
			"FROM events "+
			"WHERE event = '$pageview' "+
			"AND timestamp > now() - interval %d day "+
			"GROUP BY url "+
			"ORDER BY views DESC "+
			"LIMIT %d",
		cmd.Days, cmd.Limit,
	)

	data, err := client.Query(ctx, hogql)
	if err != nil {
		return fmt.Errorf("query top pages: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return writeRawJSON(os.Stdout, data)
	}

	return formatQueryResults(ctx, data)
}

type WebScrollDepthCmd struct {
	URL  string `required:"" help:"URL path to check scroll depth for"`
	Days int    `help:"Number of days to look back" default:"30"`
}

func (cmd *WebScrollDepthCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	hogql := fmt.Sprintf(
		"SELECT "+
			"properties.$current_url as url, "+
			"avg(toFloat64OrNull(properties.$scroll_depth)) as avg_scroll_depth, "+
			"max(toFloat64OrNull(properties.$scroll_depth)) as max_scroll_depth, "+
			"count() as pageviews "+
			"FROM events "+
			"WHERE event = '$pageview' "+
			"AND properties.$current_url LIKE '%%%s%%' "+
			"AND timestamp > now() - interval %d day "+
			"GROUP BY url "+
			"ORDER BY pageviews DESC "+
			"LIMIT 50",
		escapeHogQL(cmd.URL), cmd.Days,
	)

	data, err := client.Query(ctx, hogql)
	if err != nil {
		return fmt.Errorf("query scroll depth: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return writeRawJSON(os.Stdout, data)
	}

	return formatQueryResults(ctx, data)
}

type WebEngagementCmd struct {
	Days int `help:"Number of days to look back" default:"30"`
}

func (cmd *WebEngagementCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	hogql := fmt.Sprintf(
		"SELECT "+
			"properties.$current_url as url, "+
			"countIf(event = '$rageclick') as rage_clicks, "+
			"countIf(event = '$dead_click') as dead_clicks, "+
			"countIf(event = '$autocapture') as total_clicks, "+
			"count() as total_events "+
			"FROM events "+
			"WHERE event IN ('$rageclick', '$dead_click', '$autocapture') "+
			"AND timestamp > now() - interval %d day "+
			"GROUP BY url "+
			"ORDER BY rage_clicks DESC "+
			"LIMIT 50",
		cmd.Days,
	)

	data, err := client.Query(ctx, hogql)
	if err != nil {
		return fmt.Errorf("query engagement: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return writeRawJSON(os.Stdout, data)
	}

	return formatQueryResults(ctx, data)
}

type WebSourcesCmd struct {
	Days  int `help:"Number of days to look back" default:"30"`
	Limit int `help:"Maximum number of results" default:"20"`
}

func (cmd *WebSourcesCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	hogql := fmt.Sprintf(
		"SELECT "+
			"properties.$referring_domain as source, "+
			"count() as visits, "+
			"count(DISTINCT person_id) as unique_visitors "+
			"FROM events "+
			"WHERE event = '$pageview' "+
			"AND timestamp > now() - interval %d day "+
			"AND properties.$referring_domain IS NOT NULL "+
			"AND properties.$referring_domain != '' "+
			"GROUP BY source "+
			"ORDER BY visits DESC "+
			"LIMIT %d",
		cmd.Days, cmd.Limit,
	)

	data, err := client.Query(ctx, hogql)
	if err != nil {
		return fmt.Errorf("query sources: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return writeRawJSON(os.Stdout, data)
	}

	return formatQueryResults(ctx, data)
}
