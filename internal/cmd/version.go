package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/builtbyrobben/posthog-cli/internal/outfmt"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

type VersionCmd struct{}

func (cmd *VersionCmd) Run(ctx context.Context) error {
	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"version": VersionString(),
			"commit":  commit,
			"date":    date,
			"os":      runtime.GOOS + "/" + runtime.GOARCH,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"VERSION", "COMMIT", "DATE", "OS"}
		rows := [][]string{{VersionString(), commit, date, runtime.GOOS + "/" + runtime.GOARCH}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("posthog-cli %s\n", VersionString())
	fmt.Printf("  Commit: %s\n", commit)
	fmt.Printf("  Built:  %s\n", date)
	fmt.Printf("  OS:     %s/%s\n", runtime.GOOS, runtime.GOARCH)

	return nil
}

func VersionString() string {
	if version == "dev" {
		return "dev (no version)"
	}

	return version
}
