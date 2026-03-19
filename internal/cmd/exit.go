package cmd

import "fmt"

// ExitError wraps an error with an exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return fmt.Sprintf("exit code %d", e.Code)
}

func (e *ExitError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

// ExitCode returns the exit code for the error, or 1 as default.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitErr, ok := err.(*ExitError); ok {
		return exitErr.Code
	}

	return 1
}
