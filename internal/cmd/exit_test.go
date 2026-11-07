package cmd

import (
	"fmt"
	"testing"
)

func TestExitCodeWrapped(t *testing.T) {
	err := fmt.Errorf("command: %w", &ExitError{Code: 23})
	if got := ExitCode(err); got != 23 {
		t.Fatalf("ExitCode = %d, want 23", got)
	}
}
