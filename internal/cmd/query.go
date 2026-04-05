package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Robben-Media/posthog-cli/internal/outfmt"
)

type QueryCmd struct {
	SQL string `arg:"" required:"" help:"HogQL query to execute"`
}

func (cmd *QueryCmd) Run(ctx context.Context) error {
	client, err := getPostHogClient()
	if err != nil {
		return err
	}

	data, err := client.Query(ctx, cmd.SQL)
	if err != nil {
		return fmt.Errorf("execute query: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return writeRawJSON(os.Stdout, data)
	}

	return formatQueryResults(ctx, data)
}
